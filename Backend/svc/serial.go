package svc

import (
	"fmt"
	"time"

	"go.bug.st/serial"

	"tmaxsrv/log"
)

const (
	SEND_CH_SIZE  = 100
	RECV_CH_SIZE  = 100
	MIN_PACK_SIZE = 3
	TMP_BUF_SIZE  = 512
	QUEUE_SIZE    = 1024
)

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
type packPickerFn func(buf []byte, dataLen int) (packOffset uint, packLen uint, shouldRemoveLen uint)

type TSerial struct {
	// com port
	port serial.Port
	// send to scale channel, message will be json string
	sendCh   chan []byte
	recvCh   chan []byte
	queue    *CircularBuffer
	pickerFn packPickerFn
	baud     int    // for calculating time out
	tmpbuf   []byte // for storing temporary data read from scale
	toQuit   bool   // for informing the read/write goroutine to quit
}

// type ComInfo struct {
// 	DevPath  string // device path of Com port, e.g. COM3
// 	Baud     int    // e.x. 9600
// 	DataBits int    // value: 7,8,9
// 	StopBits int    // 0: 1 stop bit, 1: 1.5 stop bits, 2: 2 stop bits
// 	Parity   int    // 0: no parity, 1: odd, 2: even
// 	Flow     int    // 0: no flow control, 1: SW on/off flow control 2. HW CTS/RTS flow control
// }

// NewScale creates a new scale
func NewSerial(pconf ComInfo, pickerFn packPickerFn) (*TSerial, error) {
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
		port:     port,
		sendCh:   make(chan []byte, SEND_CH_SIZE),
		recvCh:   make(chan []byte, RECV_CH_SIZE),
		queue:    NewCircularBuffer(QUEUE_SIZE),
		pickerFn: pickerFn,
		baud:     pconf.Baud,
		tmpbuf:   make([]byte, TMP_BUF_SIZE),
		toQuit:   false,
	}

	go s.write()
	go s.read()

	return s, nil
}

func (s *TSerial) Close() error {
	// inform read/write goroutines to quit
	s.toQuit = true
	time.Sleep(2 * time.Millisecond) // to let read/write goroutines to quit
	// close recv/write channels

	if !IsClosed(s.recvCh) {
		close(s.recvCh)
	}

	if !IsClosed(s.sendCh) {
		close(s.sendCh)
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

func (s *TSerial) Write(data []byte) error {
	if len(s.sendCh) >= SEND_CH_SIZE {
		log.Log.Error("sendCh is full")
		return errFull
	}
	s.sendCh <- data
	return nil
}

// read messages from the scale and parses them into data record or command response
// The application runs read in a per-scale goroutine. The application
// ensures that there is at most one reader on a scale by executing all
// reads from this goroutine.
func (c *TSerial) read() {
	packCnt := 0
	//readTimeOut := time.Duration(float64(256)*1000/float64(c.baud)) * time.Millisecond // read 256 bytes time
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
			continue
		} else if n == 0 {
			time.Sleep(1 * time.Millisecond) // to avoid consume too much cpu time
		}
		if c.queue.GetDataLen() > MIN_PACK_SIZE {
			// call the packet picker function
			data := c.queue.PeekAll()
			packOff, packLen, removeLen := c.pickerFn(data, c.queue.GetDataLen())
			if packLen > 0 {
				if len(c.recvCh) >= RECV_CH_SIZE {
					log.Log.Errorf("recvCh full, size: %v", len(c.recvCh))
				} else {
					c.recvCh <- data[packOff : packOff+packLen]
				}
				packCnt++
			} else {
				// fmt.Printf("no packet\n")
			}
			if removeLen > 0 {
				c.queue.DequeueN(int(removeLen))
			}
		}
		// time.Sleep(1 * time.Millisecond)
	}
}

func (s *TSerial) readScale() (int, error) {
	n, err := s.port.Read(s.tmpbuf)
	if err != nil {
		log.Log.Errorf("Error reading scale: %v", err)
		return 0, err
	}
	//fmt.Printf("data:%v", string(tmpBuf))

	if s.queue.IsFull() { // FIXME: should we handle this error with this way?
		s.queue.DequeueN(s.queue.capacity) // handle abnormal case
	}

	if n > 0 {
		if err := s.queue.EnqueueN(s.tmpbuf[0:n], n); err != nil {
			s.queue.Reset()
		}
	}

	return n, nil
}

// A goroutine running write is started for the scale. The
func (s *TSerial) write() {
	for {
		if s.toQuit {
			break // quit immediately
		}
		select {
		case message, ok := <-s.sendCh:
			if !ok {
				// The hub closed the channel.
				return
			}
			// send message to scale
			if s.port != nil {
				n, err := s.port.Write(message)
				if err != nil || n != len(message) {
					log.Log.Error("Error on sending message to scale, to send: %v, sent:%v, err:%v\n", len(message), n, err.Error())
				}
			}
		}
	}
}

func comInfo2SerialMode(pcnf ComInfo) *serial.Mode {
	var mode serial.Mode
	mode.BaudRate = pcnf.Baud
	mode.DataBits = pcnf.DataBits
	if pcnf.Parity == 0 {
		mode.Parity = serial.NoParity
	} else if pcnf.Parity == 1 {
		mode.Parity = serial.EvenParity
	} else if pcnf.Parity == 2 {
		mode.Parity = serial.OddParity
	} else {
		mode.Parity = serial.NoParity
	}
	if pcnf.StopBits == 0 {
		mode.StopBits = serial.OneStopBit
	} else if pcnf.StopBits == 1 {
		mode.StopBits = serial.OnePointFiveStopBits
	} else if pcnf.StopBits == 2 {
		mode.StopBits = serial.TwoStopBits
	} else {
		mode.StopBits = serial.OneStopBit
	}
	return &mode
}
