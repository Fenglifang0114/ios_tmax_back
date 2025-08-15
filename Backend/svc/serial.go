package svc

import (
	"fmt"
	"sync"
	"time"

	"go.bug.st/serial"

	"tmaxsrv/comm"
	"tmaxsrv/log"
	"tmaxsrv/picker"
	"tmaxsrv/util"
)

const (
	SEND_CH_SIZE  = 1000
	RECV_CH_SIZE  = 1000
	MIN_PACK_SIZE = 3
	TMP_BUF_SIZE  = 102400
	QUEUE_SIZE    = 102400
)

// var RESP_SERIAL_ERROR = []byte{0x5a, 0xa5, 0x00, 0x01, 0x7f, 0xBD, 0x86, 0x1C, 0x86, 0xa5, 0x5a}
var RESP_SERIAL_ERROR = []byte{0x5a, 0xa5, 0x00, 0x0B, 0xff, 0x12, 0x00, 0xD9, 0x02, 0x8D, 0x99, 0xa5, 0x5a}

// this function should try to scan the input buffer to find the first packet,
// and return the package start position, the package len and how many bytes should be removed from the buffer,
// if no package is found then the returned packLen should be zero and the shouldRemoveLen may be not zero,
// that means no use data in the input buffer.
// buf: the data should be scanned to pick the package
// dataLen: the length of the data in the buffer
// return:
// packageOffset: the package start position
// packLen: the number of bytes in the package that scanned
// shouldRemoveLen: the number of bytes should be removed, containing the no use data in the buffer
// type packPickerFn func(buf []byte, dataLen int) (packOffset uint, packLen uint, shouldRemoveLen uint, pack comm.Packet)

type TSerial struct {
	// com port
	port serial.Port
	// send to scale channel, message will be json string
	sendCh     chan []byte
	recvCh     chan comm.Packet
	wg         sync.WaitGroup
	queue      *util.CircularBuffer
	pickerFn   picker.PickerFunc
	baud       int    // for calculating time out
	tmpbuf     []byte // for storing temporary data read from scale
	toQuit     bool   // for informing the read/write goroutine to quit
	maxPackLen int    // max package size that scale can receive on time
	isDefault  bool
}

// NewSerial creates a new serial port
func NewSerial(pconf ComInfo, pickerFn picker.PickerFunc, isDefault bool) (*TSerial, error) {
	if pickerFn == nil {
		return nil, fmt.Errorf("user NewSerial(), should provide a picker function")
	}
	var port serial.Port
	var err error
	mode := comInfo2SerialMode(pconf)

	port, err = serial.Open(pconf.DevPath, mode)

	log.Log.Infof("open port:%v with Mode: %v", pconf.DevPath, mode)
	if err != nil {
		log.Log.Errorf("error on open port:%v, error: %v", pconf.DevPath, err)
		port = nil
	} else {
		port.SetReadTimeout(20 * time.Millisecond)
	}
	s := &TSerial{
		port:      port,
		sendCh:    make(chan []byte, SEND_CH_SIZE),
		recvCh:    make(chan comm.Packet, RECV_CH_SIZE),
		queue:     util.NewCircularBuffer(QUEUE_SIZE),
		pickerFn:  pickerFn,
		baud:      pconf.Baud,
		tmpbuf:    make([]byte, TMP_BUF_SIZE),
		toQuit:    false,
		isDefault: isDefault,
	}

	go s.write()
	go s.read()

	return s, nil
}

func (s *TSerial) Close() error {
	// inform read/write goroutines to quit
	s.toQuit = true
	time.Sleep(10 * time.Millisecond) // to let goroutines run

	// waiting s.read() goroutine to quit
	s.wg.Wait()

	// close recv/write channels
	if !IsClosed(s.sendCh) {
		close(s.sendCh)
	}
	// wait for goroutine quit
	if !IsPacketChClosed(s.recvCh) {
		close(s.recvCh)
	}

	// close serial port
	if s.port == nil {
		return fmt.Errorf("s.comPort is nil")
	}
	log.Log.Info("close serial port")
	s.port.Close()
	time.Sleep(50 * time.Millisecond) // to let port closed
	return nil
}

func IsClosed(ch <-chan []byte) bool {
	select {
	case <-ch:
		return true
	default:
	}

	return false
}

func IsPacketChClosed(ch <-chan comm.Packet) bool {
	select {
	case <-ch:
		return true
	default:
	}

	return false
}

func (s *TSerial) Write(data []byte) error {
	if len(s.sendCh) >= SEND_CH_SIZE {
		log.Log.Error("sendCh is full")
		return util.ErrFull
	}
	s.sendCh <- data
	return nil
}

// read messages from the scale and parses them into data record or command response
// The application runs read in a per-scale goroutine. The application
// ensures that there is at most one reader on a scale by executing all
// reads from this goroutine.
func (c *TSerial) read() {
	c.wg.Add(1)
	packCnt := 0
	// readTimeOut := time.Duration(float64(256)*1000/float64(c.baud)) * time.Millisecond // read 256 bytes time
	for {
		if c.toQuit {
			break // quit immediately
		}

		if c.port == nil {
			time.Sleep(10 * time.Millisecond) // to avoid consume too much cpu time
			continue
		}

		// read data from serial at least PACK_MIN_LEN or timeout (2 * 1/baud)
		if n, err := c.readScale(); err != nil { // data will be stored in the queue
			log.Log.Errorf("@TSerial read(), err: %v\n", err)
			if !IsPacketChClosed(c.recvCh) {
				c.recvCh <- comm.Packet{PayloadLen: uint16(len(RESP_SERIAL_ERROR)), CmdID: 0xff, CmdSubId: 0x12, SeqNum: 0, Payload: []byte{}}
			}
			time.Sleep(10 * time.Second) // to avoid sending error too often to UI
			continue
		} else if n == 0 {
			time.Sleep(1 * time.Millisecond) // to avoid consume too much cpu time
		}
		if c.queue.GetDataLen() > MIN_PACK_SIZE {
			// call the packet picker function
			data := c.queue.PeekAll()
			_, packLen, removeLen, pack := c.pickerFn(data, c.queue.GetDataLen())
			if packLen > 0 {
				if len(c.recvCh) >= RECV_CH_SIZE {
					log.Log.Errorf("recvCh full, size: %v", len(c.recvCh))
					fmt.Printf("recvCh full, size: %v", len(c.recvCh))
				} else {
					c.recvCh <- pack
				}
				packCnt++
			} else {
				// fmt.Printf("no packet\n")
			}
			if removeLen > 0 {
				c.queue.DequeueN(int(removeLen))
			}
		}
	}
	c.wg.Done()
}

func (s *TSerial) readScale() (int, error) {
	n, err := s.port.Read(s.tmpbuf)
	if err != nil {
		log.Log.Errorf("Error reading scale: %v", err)

		return 0, err
	}
	// fmt.Printf("data:%v", string(tmpBuf))

	if s.queue.IsFull() {
		s.queue.DequeueN(s.queue.Capacity) // handle abnormal case
	}

	if n > 0 {

		log.Log.Infof("******serial port Got: %x ", string(s.tmpbuf[0:n]))

		log.Log.Infof("******serial port Got: %v ", string(s.tmpbuf[0:n]))
		// fmt.Printf("recv data:%x ", string(s.tmpbuf[0:n]))
		// fmt.Printf("data:%s\n", string(s.tmpbuf[0:n]))
		if err := s.queue.EnqueueN(s.tmpbuf[0:n], n); err != nil {
			s.queue.Reset()
		}
	}

	return n, nil
}

// A goroutine running write is started for the scale. The
func (s *TSerial) write() {
	for message := range s.sendCh {
		// send message to scale
		if s.port != nil && s.isDefault {
			n, err := s.port.Write(message)
			// fmt.Printf("out:%x\n", message)
			// fmt.Printf("out:%s\n", string(message))
			if err != nil || n != len(message) {
				log.Log.Error(fmt.Sprintf("Error on sending message to scale, to send: %v, sent:%v, err:%v\n", len(message), n, err.Error()))
			}
		}
	}
}

func (s *TSerial) ChangePickFunc(pickerFn picker.PickerFunc) {
	s.pickerFn = pickerFn
}

func comInfo2SerialMode(pcnf ComInfo) *serial.Mode {
	var mode serial.Mode
	mode.BaudRate = pcnf.Baud
	mode.DataBits = pcnf.DataBits
	// 将Parity的if-else转换为switch
	switch pcnf.Parity {
	case 1:
		mode.Parity = serial.EvenParity
	case 2:
		mode.Parity = serial.OddParity
	default:
		mode.Parity = serial.NoParity
	}
	// 将StopBits的if-else转换为switch
	switch pcnf.StopBits {
	case 1:
		mode.StopBits = serial.OnePointFiveStopBits
	case 2:
		mode.StopBits = serial.TwoStopBits
	default:
		mode.StopBits = serial.OneStopBit
	}
	return &mode
}
