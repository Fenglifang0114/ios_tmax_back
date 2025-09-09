package svc

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"tmaxsrv/comm"
	"tmaxsrv/log"
	"tmaxsrv/picker"
	"tmaxsrv/util"
)

type TNet struct {
	// net port
	port int
	ip   string
	// conn net.Conn
	conn *net.TCPConn
	// send to scale channel, message will be json string
	sendCh     chan []byte
	recvCh     chan comm.Packet
	wg         sync.WaitGroup
	queue      *util.CircularBuffer
	pickerFn   picker.PickerFunc
	tmpbuf     []byte // for storing temporary data read from scale
	toQuit     bool   // for informing the read/write goroutine to quit
	maxPackLen int    // max package size that scale can receive on time
	isAlive    bool
	isDefault  bool //用来判断同一台秤，多种连接方式的情况下，是否走这个通道
}

// NewSerial creates a new serial port

func NewNet(ncnf NetInfo, pickerFn picker.PickerFunc, isDefault bool) (*TNet, error) {
	if pickerFn == nil {
		return nil, fmt.Errorf("user NewNet(), should provide a picker function")
	}
	ip := ncnf.Ip
	port := ncnf.Port
	var conn *net.TCPConn = nil
	connected := false
	tnet := &TNet{
		port:      port,
		ip:        ip,
		conn:      conn,
		sendCh:    make(chan []byte, SEND_CH_SIZE),
		recvCh:    make(chan comm.Packet, RECV_CH_SIZE),
		queue:     util.NewCircularBuffer(QUEUE_SIZE),
		pickerFn:  pickerFn,
		tmpbuf:    make([]byte, TMP_BUF_SIZE),
		toQuit:    false,
		isAlive:   connected,
		isDefault: isDefault,
	}

	go tnet.write()
	go tnet.read()

	return tnet, nil
}

func (tnet *TNet) Close() error {
	// inform read/write goroutines to quit
	tnet.toQuit = true
	println("toQuit==============")
	// time.Sleep(100000 * time.Millisecond) // to let goroutines run
	time.Sleep(100 * time.Millisecond) // to let goroutines run
	if tnet.conn != nil {
		err := tnet.conn.Close()
		if err != nil {
			return fmt.Errorf("tnet.conn closed error")
		}
	}

	// waiting s.read() goroutine to quit
	fmt.Println("关闭 conn")
	tnet.wg.Wait() //20240801
	fmt.Println("关闭 read")
	// close recv/write channels
	if !IsClosed(tnet.sendCh) {
		close(tnet.sendCh)
	}
	// wait for goroutine quit
	if !IsPacketChClosed(tnet.recvCh) {
		close(tnet.recvCh)
	}

	if !tnet.toQuit {
		tnet.toQuit = true
	}

	// close tcp connect

	if tnet.conn == nil {
		return fmt.Errorf("tnet.conn is nil")
	}
	log.Log.Info("close tcp connect")

	// err = tnet.conn.Close()
	// if err != nil {
	// 	return fmt.Errorf("tnet.conn closed error")
	// }

	time.Sleep(50 * time.Millisecond) // to let tnet closed
	return nil
}

func (tnet *TNet) Write(data []byte) error {
	if len(tnet.sendCh) >= SEND_CH_SIZE {
		log.Log.Error("sendCh is full")
		return util.ErrFull
	}
	tnet.sendCh <- data
	return nil
}

// read messages from the scale and parses them into data record or command response
// The application runs read in a per-scale goroutine. The application
// ensures that there is at most one reader on a scale by executing all
// reads from this goroutine.
func (tnet *TNet) read() {
	tnet.wg.Add(1)
	packCnt := 0

	for {
		if tnet.toQuit {
			break // quit immediately
		}

		if tnet.conn == nil {
			time.Sleep(10 * time.Millisecond) // to avoid consume too much cpu time
			continue
		}

		// read data from net at least PACK_MIN_LEN or timeout (2 * 1/baud)
		if n, err := tnet.readScale(); err != nil { // data will be stored in the queue
			log.Log.Errorf("@TNet read(), err: %v\n", err)
			if tnet.toQuit {
				continue
			}

			if err.Error() != "EOF" {
				println("断开连接，重新连")
				tnet.conn.Close()
				tnet.conn = nil
				continue
			}

			if err.Error() == "EOF" { // 20240801
				log.Log.Errorf("EOF")
				//20250901 断开TCP，重新连接
				tnet.conn.Close()
				tnet.reconnect()
				continue
			}
			if !IsPacketChClosed(tnet.recvCh) {
				// tnet.recvCh <- comm.Packet{PayloadLen: uint16(len(RESP_SERIAL_ERROR)), CmdID: 0, CmdSubId: 0, SeqNum: 0, Payload: RESP_SERIAL_ERROR}
			}
			time.Sleep(100 * time.Millisecond) // to avoid sending error too often to UI
			continue
		} else if n == 0 {
			time.Sleep(1 * time.Millisecond) // to avoid consume too much cpu time
		}
		hasPack := true
		for hasPack {
			if tnet.queue.GetDataLen() > MIN_PACK_SIZE {
				// call the packet picker function
				data := tnet.queue.PeekAll()
				_, packLen, removeLen, pack := tnet.pickerFn(data, tnet.queue.GetDataLen())
				if packLen > 0 {
					if len(tnet.recvCh) >= RECV_CH_SIZE {
						log.Log.Errorf("recvCh full, size: %v", len(tnet.recvCh))
						fmt.Printf("recvCh full, size: %v", len(tnet.recvCh))
					} else {
						// fmt.Printf("pack data: %v\n", pack.Payload)
						tnet.recvCh <- pack

					}
					packCnt++
				} else {
					fmt.Printf("no packet\n")
					hasPack = false
				}
				if removeLen < packLen {
					fmt.Printf("removelen error\n")
				}
				if removeLen > 0 {
					tnet.queue.DequeueN(int(removeLen))
				}
			} else {
				hasPack = false
			}

		}

	}
	tnet.wg.Done()

}

// if n > 0 {
// 	hexData := hex.EncodeToString(tnet.tmpbuf[:n])
// 	fmt.Printf("Received HEX data: %s\n", hexData)
// }

func (tnet *TNet) readScale() (int, error) {
	if !tnet.isAlive {
		time.Sleep(1 * time.Second)
		return 0, nil //关闭了会报错
	}
	if tnet.toQuit {
		time.Sleep(1 * time.Second)
		return 0, nil //关闭了会报错

	}

	n, err := tnet.conn.Read(tnet.tmpbuf)
	if err != nil {
		log.Log.Errorf("Error reading scale: %v", err)
		tnet.isAlive = false
		return 0, err
	}

	if tnet.queue.IsFull() {
		tnet.queue.DequeueN(tnet.queue.Capacity) // handle abnormal case
	}

	if n > 0 {
		// log.Log.Debug(tnet.tmpbuf[0:n])
		fmt.Printf("netrevdata:%x\n", string(tnet.tmpbuf[0:n]))
		fmt.Printf("netrevdata:%s\n", string(tnet.tmpbuf[0:n]))
		if err := tnet.queue.EnqueueN(tnet.tmpbuf[0:n], n); err != nil {
			tnet.queue.Reset()
		}
	}

	return n, nil
}

// A goroutine running write is started for the scale. The
func (tnet *TNet) write() {
	for message := range tnet.sendCh {
		// send message to scale
		if tnet.conn != nil {
			n, err := tnet.conn.Write([]byte(message))

			if err != nil || n != len(message) {
				tnet.isAlive = false
				log.Log.Error(fmt.Sprintf("Error on sending message to scale, to send: %v, sent:%v, err:%v\n", len(message), n, err.Error()))

			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func (tnet *TNet) ChangePickFunc(pickerFn picker.PickerFunc) {
	tnet.pickerFn = pickerFn
	//TODO:  修改网络信息  待写
}

// func (tnet *TNet) monitorConnection() {
// 	ticker := time.NewTicker(2 * time.Second)
// 	for range ticker.C {
// 		if tnet.conn == nil || !tnet.isAlive {
// 			// 尝试重连
// 			var err error
// 			tnet.conn, err = tnet.reconnect()
// 			if err != nil {
// 				log.Log.Printf("Reconnection failed: %v", err)
// 			}
// 		} else {
// 			fmt.Println("Connection is alive")
// 		}
// 	}
// }

func (tnet *TNet) reconnect() (*net.TCPConn, error) {

	conn, err := tnet.connect()
	if err == nil {
		tnet.conn = conn
		tnet.isAlive = true
	}
	println("reconnect")
	println(tnet.ip)

	return conn, err
}

func (tnet *TNet) connect() (*net.TCPConn, error) {

	conn, err := tnet.dialTCP()
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (tnet *TNet) dialTCP() (*net.TCPConn, error) {

	tcpAddr, err := net.ResolveTCPAddr("tcp", tnet.ip+":"+strconv.Itoa(tnet.port))
	if err != nil {
		log.Log.Printf("ResolveTCPAddr error: %v", err)
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)

	// conn, err := net.Dial("tcp", tnet.ip+":"+strconv.Itoa(tnet.port))

	if err != nil {
		return nil, err
	}
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(1 * time.Second)

	return conn, nil

}
