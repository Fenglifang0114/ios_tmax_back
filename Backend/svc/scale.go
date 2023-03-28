package svc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"

	"tmaxsrv/comm"
	"tmaxsrv/log"
	"tmaxsrv/prnfmt"
	utils "tmaxsrv/utils"
)

var GlastWantRespMsgType RespMsgType = NO_RESP

const (
	SCALE_RECV_CH_SIZE = 16
	SCALE_SEND_CH_SIZE = 16
)

const (
	GET_WEIGHT_CMD string = "W\r\n"
	TARE_CMD       string = "T\r\n"
	ZERO_CMD       string = "Z\r\n"
)

const (
	CMD_HEAD1 byte = 0x5A
	CMD_HEAD2 byte = 0xA5
	CMD_TAIL1 byte = 0xA5
	CMD_TAIL2 byte = 0x5A

	HEADER_LEN int = 2

	CMD_CONF_TYPE           byte = 0x80
	CMD_CAL0_TYPE           byte = 0x81
	CMD_CALN_TYPE           byte = 0x82
	CMD_LOCK_TYPE           byte = 0x83
	CMD_ZERO_TYPE           byte = 0x84
	CMD_TARE_TYPE           byte = 0x85
	CMD_R1_TYPE             byte = 0x86
	CMD_R1_CONT_TYPE        byte = 0x87
	CMD_R10_TYPE            byte = 0x89
	CMD_R10_CONT_TYPE       byte = 0x8A
	CMD_CODE_TYPE           byte = 0x8C
	CMD_CODE_CONT_TYPE      byte = 0x8D
	CMD_CODE_CONT_STOP_TYPE byte = 0x8E
	RECV_CODE_TYPE          byte = 0xA2

	CMD_PACK_NO_DATA_LEN = 8

	CMD_LEN_IDX = 2
	CMD_SEQ_IDX = 5

	DUMMY_SEQ byte = 0
)

const (
	SCALE_TIME_OUT_S = 10 // 10 seconds
)

type Scale struct {
	scaleMgr *ScaleMgr
	// connectivity Media interface
	Conn *ScaleConnMedia
	// media config, this is a json string that will be unmarshaled to specific structure typed value
	//MediaConf string
	// com port
	MySerial *TSerial
	// Network socket
	//tcpSocket net.Socket
	// Bluetooth
	//btConn BtCom
	// scale Id
	Id int64
	// scale Model, if not supported then the default Model is "legacy"
	Model string
	// scale serial number, if not supported then the default serial number is "123456789"
	Sn string
	// send to scale channel, message will be json string
	toScaleMsgCh chan string
	// receive from scale channel, message will be json string
	fromScaleMsgCh chan string
	// added time
	EnterAt       time.Time
	respChans     map[RespMsgType][]chan *ScaleRespMsg
	isOldC51Scale bool
	// if client needs unsolicited data from scale
	isSendUnolicitedData bool
	// to scale command is issued and waiting response
	isWaintingResp bool
	// client that will communicate with scale
	client *Client
	// this channel is to inform the procScaleRespMessage goroutine, which process the data from serial port, to quit
	quitProcScaleRespMessageCh chan struct{}
	// this channel is to inform the procToScaleMsg, which process the data from serial port, to quit
	quitProcToScaleMsgCh chan struct{}
}

// NewScale creates a new scale
func NewScale(scaleMgr *ScaleMgr, conn *ScaleConnMedia, model string, sn string, isTest bool) (*Scale, error) {
	var pcnf ComInfo
	var sport *TSerial
	var err error
	if conn.TMedia == MEDIA_COM {
		// open COM connection
		if err := json.Unmarshal([]byte(conn.MediaConf.MediaInfoJson), &pcnf); err != nil {
			log.Log.Errorf("error unmarshalling: %v", err)
		}
		// mode := mediaInfoToMode(pcnf)
		if comm.GcurScale == comm.SCALE_TYPE_TMAX {
			sport, err = NewSerial(pcnf, pickerFnTmaxScale)
			if err != nil {
				sport = nil
			}
		} else if comm.GcurScale == comm.SCALE_TYPE_OLD_C51 {
			sport, err = NewSerial(pcnf, pickerFnOldScale)
			if err != nil {
				sport = nil
			}
		} else {
			log.Log.Error("not support scale type")
			return nil, fmt.Errorf("not support scale type")
		}
	}
	scale := &Scale{scaleMgr: scaleMgr, Conn: conn, Model: model, Sn: sn, toScaleMsgCh: make(chan string, SCALE_SEND_CH_SIZE),
		fromScaleMsgCh: make(chan string, SCALE_RECV_CH_SIZE), MySerial: sport,
		quitProcScaleRespMessageCh: make(chan struct{}), quitProcToScaleMsgCh: make(chan struct{})}
	scale.respChans = map[RespMsgType][]chan *ScaleRespMsg{}
	scale.respChans[WEIGHT_DATA] = []chan *ScaleRespMsg{}
	scale.respChans[ZERO_CMD_RESP] = []chan *ScaleRespMsg{}
	scale.respChans[TARE_CMD_RESP] = []chan *ScaleRespMsg{}
	scale.respChans[WEIGHT_DATA_RESP] = []chan *ScaleRespMsg{}
	scale.respChans[DOWN_PRN_FMT_RESP] = []chan *ScaleRespMsg{}

	scale.isOldC51Scale = true // in this moment we use this software for old C51 scales, for T-Max scales should set this to false

	scale.isSendUnolicitedData = false
	scale.isWaintingResp = false

	scale.scaleMgr = scaleMgr
	scale.EnterAt = time.Now()
	scale.MySerial = sport
	go scale.procScaleRespMsg()
	go scale.procToScaleMsg()
	return scale, nil
}

func (s *Scale) Close() error {
	close(s.quitProcScaleRespMessageCh)
	close(s.quitProcToScaleMsgCh)
	if s.MySerial != nil {
		s.MySerial.Close()
	}
	return nil
}

func (s *Scale) SetClient(client *Client) error {
	s.client = client
	return nil
}

func (s *Scale) HandleClientDisconnect() error {
	log.Log.Warn("Client disconnected, HandleClientDisconnect called")
	s.client = nil
	return nil
}

// to process msg from serial port
func (s *Scale) procScaleRespMsg() {
	quit := false
	for {
		if quit {
			break
		}
		select {
		case _ = <-s.quitProcScaleRespMessageCh:
			quit = true
		case inData := <-s.MySerial.recvCh:
			if inData == nil { // TODO: maybe caused by closed serial port
				continue
			}
			log.Log.Debugf("From sport: %v", inData)
			if !s.isSendUnolicitedData || !s.isWaintingResp {
				// continue // FIXME: skip this line for testing purpose
			}
			if comm.GcurScale == comm.SCALE_TYPE_OLD_C51 {
				if msg, err := extractMsgOldScale(inData); msg != nil && err == nil {
					if len(s.fromScaleMsgCh) < RECV_CH_SIZE {
						if msgStr, err := json.MarshalToString(msg); err == nil {
							if s.client != nil {
								s.client.sendCh <- []byte(msgStr)
							}
							//s.fromScaleMsgCh <- msgStr
						} else {
							log.Log.Errorf("marshal msg err: %v", err.Error())
						}
					} else {
						log.Log.Error("fromScaleMsgCh full")
					}
				} else {
					if msg != nil { // it's error
						log.Log.Errorf("extractMsgOldScale error: %v", err.Error())
					}
				}
			} else if comm.GcurScale == comm.SCALE_TYPE_TMAX {
				if msg, err := extractMsgTmaxScale(s.Id, inData); msg != nil && err == nil {
					chs := s.respChans[msg.MsgType]
					if chs != nil {
						for _, ch := range chs {
							if len(ch) == 0 { // to avoid blocking, this kind of channel should only be used once
								ch <- msg
							}
						}
					}
					if len(s.fromScaleMsgCh) < RECV_CH_SIZE { // no use in this moment
						if msgStr, err := json.MarshalToString(msg); err == nil {
							if s.client != nil {
								s.client.sendCh <- []byte(msgStr)
							}
							//s.fromScaleMsgCh <- msgStr
						} else {
							log.Log.Errorf("marshal msg err: %v", err.Error())
						}
					} else {
						log.Log.Error("fromScaleMsgCh full")
					}
				} else {
					if msg != nil { // it's error
						log.Log.Errorf("extractMsgTmaxScale error: %v", err.Error())
					}
				}
			}
		}
	}
}

func (s *Scale) procToScaleMsg() {
	quit := false
	for {
		if quit {
			break
		}
		if s.client == nil {
			time.Sleep(10 * time.Millisecond) // to avoid consume too many cpu resource
			continue
		}
		select {
		case _ = <-s.quitProcToScaleMsgCh:
			quit = true
		case userMessage, ok := <-s.client.recvCh:
			if !ok {
				log.Log.Error("client's recvCh closed")
				continue
			}
			var data map[string][]byte
			json.Unmarshal(userMessage, &data)
			log.Log.Debugf("userMessage: %v", data)
			scaleId := new(big.Int).SetBytes(data["scaleId"]).Int64()
			if scaleId != s.Id {
				log.Log.Errorf("wrong scale id :%v received, our Id is: %v", scaleId, s.Id)
				continue
			}
			if data == nil { // TODO: maybe caused by ...client?
				continue
			}
			log.Log.Debugf("From wsclient: %v", string(data["message"]))
			if req, err := parseToScaleReq(string(data["message"])); err == nil {
				go procToScaleReq(s, req) // TODO: handle error
			}
		}
	}
}

// func mediaInfoToMode(pcnf ComInfo) *serial.Mode {
// 	var mode serial.Mode
// 	mode.BaudRate = pcnf.Baud
// 	mode.DataBits = pcnf.DataBits
// 	if pcnf.Parity == 0 {
// 		mode.Parity = serial.NoParity
// 	} else if pcnf.Parity == 1 {
// 		mode.Parity = serial.EvenParity
// 	} else if pcnf.Parity == 2 {
// 		mode.Parity = serial.OddParity
// 	} else {
// 		mode.Parity = serial.NoParity
// 	}
// 	if pcnf.StopBits == 0 {
// 		mode.StopBits = serial.OneStopBit
// 	} else if pcnf.StopBits == 1 {
// 		mode.StopBits = serial.OnePointFiveStopBits
// 	} else if pcnf.StopBits == 2 {
// 		mode.StopBits = serial.TwoStopBits
// 	} else {
// 		mode.StopBits = serial.OneStopBit
// 	}
// 	return &mode
// }

func (c *Scale) ModifyMedia(conf MediaConf) bool {
	//	c.Lock()
	//	defer c.Unlock()
	if conf.Type == MEDIA_COM {
		var pcnf ComInfo
		var err error
		// open COM connection
		if err := json.Unmarshal([]byte(conf.MediaInfoJson), &pcnf); err != nil {
			log.Log.Errorf("error unmarshalling: %v", err)
		}
		if c.MySerial != nil {
			c.MySerial.Close()
		}
		if c.MySerial, err = NewSerial(pcnf, pickerFnOldScale); err != nil {
			log.Log.Error(err.Error())
		}
	} else if conf.Type == MEDIA_NET {

	} else if conf.Type == MEDIA_BT {

	} else {

	}
	return true
}

func (c *Scale) PerfZero() bool {
	if c.isOldC51Scale {
		return perfCmd(c, []byte(ZERO_CMD))
	} else {
		// TODO: for T-Max
	}
	return true
}

func (c *Scale) PerfTare() bool {
	if c.isOldC51Scale {
		return perfCmd(c, []byte(TARE_CMD))
	} else {
		// TODO: for T-Max
	}
	return true
}

func (c *Scale) ReadWeight() bool {
	if c.isOldC51Scale {
		return perfCmd(c, []byte(GET_WEIGHT_CMD))
	} else {
		// TODO: for T-Max
	}
	return true
}

func perfCmd(c *Scale, cmd []byte) bool {
	if err := writeScale(c, cmd); err != nil {
		log.Log.Error(err.Error())
		return false
	}
	return true
}

func perfCmdNwaitResult(c *Scale, cmd []byte, waitMsgType RespMsgType) (*ScaleRespMsg, error) {
	ch := make(chan *ScaleRespMsg, 10)
	c.RegisterNotif(waitMsgType, ch)
	defer func() {
		c.UnRegisterNotif(waitMsgType, ch)
		c.isWaintingResp = false
	}()

	GlastWantRespMsgType = waitMsgType
	c.isWaintingResp = true
	if err := writeScale(c, cmd); err != nil {
		log.Log.Error(err.Error())
		return &ScaleRespMsg{MsgType: NO_RESP}, err
	}

	var ret *ScaleRespMsg
	select {
	case ret = <-ch:
	case <-time.After(SCALE_TIME_OUT_S * time.Second):
		ret = &ScaleRespMsg{MsgType: NO_RESP}
	}
	if ret.MsgType == NO_RESP {
		return ret, fmt.Errorf("no response, time out")
	}

	return ret, nil
}

func (c *Scale) RegWeightData() bool {
	c.isSendUnolicitedData = true
	return true
}

func (c *Scale) UnRegWeightData() bool {
	c.isSendUnolicitedData = false
	return true
}

func (c *Scale) GetRecs() ([]ScaleRec, error) {
	return c.scaleMgr.recPb.GetRecsList(*c)
}

func (c *Scale) AddRec(rec ScaleRec) error {
	rec.ScaleModel = c.Model
	rec.ScaleSn = c.Sn
	return c.scaleMgr.recPb.InsertRec(rec)
}

func (c *Scale) DelRec(recId uint) error {
	return c.scaleMgr.recPb.DeleteRec(recId)
}

// func readScale(c *Scale) error {
// 	// TODO: read data from scale's serial port or net interface or BT interface according the media type that connecting to scale
// 	//msgLen := 0
// 	//isGotMsg := false
// 	if c.Conn.TMedia == MEDIA_COM {
// 		if c.MySerial == nil {
// 			time.Sleep(time.Millisecond * 50) // to avoid consume too much cpu time
// 			return nil
// 		}
// 		tmpBuf := make([]byte, 1024)
// 		n, err := c.MySerial.Read(tmpBuf)
// 		if err != nil {
// 			log.Log.Error("Error reading scale: %v", err)
// 			return err
// 		}
// 		//fmt.Printf("data:%v", string(tmpBuf))

// 		if c.queue.IsFull() { // FIXME: should we handle this error with this way?
// 			c.queue.DequeueN(c.queue.capacity) // handle abnormal case
// 		}

// 		if n > 0 {
// 			// c.readBuf = append(c.readBuf, tmpBuf...)
// 			for i := 0; i < n; i++ {
// 				c.queue.Enqueue(tmpBuf[i])
// 			}
// 		} else {
// 			time.Sleep(10 * time.Millisecond) // to avoid consume too much cpu time
// 		}

// 		//if c.queue.DataLen() > 20 {
// 		//	data, _ := c.queue.Peekn(0, c.queue.DataLen())
// 		//fmt.Printf(string(data))
// 		//}

// 		// if n == 0 {
// 		// 	time.Sleep(time.Millisecond * 1) // let other goroutine to grab cpu time
// 		// }
// 	} else if c.Conn.TMedia == MEDIA_NET {
// 		// TODO: get data from network
// 	} else if c.Conn.TMedia == MEDIA_BT {
// 		// TODO: get data from bluetooth
// 	} else {
// 		// something wrong
// 	}

// 	// TODO: extract message from the buffer
// 	return nil
// }

// return WeightMsg package and return the len of parsed in data
func retreiveWeightC51(data []byte) (pack WeightMsg, err error) {
	dataStr := string(data)
	// fields will be "ST,NT, 5.123kg" or "ST,NT,-0.123kg", or "-- UL --", "-- OL --"
	fields := strings.Split(dataStr, ",")
	if len(fields) != 3 {
		if len(fields[0]) < MIN_PACK_SIZE {
			return WeightMsg{}, fmt.Errorf("No packet")
		}
		return WeightMsg{WeightVal: dataStr, WeightUnit: ""}, nil
	} else {
		weightMsg := WeightMsg{}
		weightMsg.IsStable = strings.Contains(fields[0], "ST")
		weightMsg.IsNet = strings.Contains(fields[1], "NT")
		regexp, err := regexp.Compile("([0-9.-]+)([a-zA-Z]+)")
		if err != nil {
			return WeightMsg{}, err
		}
		match := regexp.FindStringSubmatch(fields[2])
		if len(match) != 3 { // 5.123kg, 5.123, kg
			return WeightMsg{}, fmt.Errorf("finding substring error: %v", fields[2])
		}
		weightMsg.WeightVal = strings.TrimSpace(match[1])
		weightMsg.WeightUnit = strings.TrimSpace(match[2])
		return weightMsg, nil
	}
}

func extractMsgOldScale(data []byte) (*ScaleRespMsg, error) {
	// TODO: check if data is response or unsolicited message, our C51 MCU scale will not response any character for T/Z/W
	msg, err := retreiveWeightC51(data)
	if err != nil { // no packet ready, just return
		log.Log.Info(err.Error())
		return nil, nil
	}

	scaleMsg := ScaleRespMsg{}
	scaleMsg.MsgType = WEIGHT_DATA
	scaleMsg.MsgBody = msg
	return &scaleMsg, err
}

func extractMsgTmaxScale(scaleId int64, data []byte) (*ScaleRespMsg, error) {
	// TODO: check if data is response or unsolicited message, our C51 MCU scale will not response any character for T/Z/W
	msg, err := retreiveRespMsg(scaleId, data)
	if err != nil { // no packet ready, just return
		log.Log.Info(err.Error())
		return nil, err
	}

	return msg, err
}

func retreiveRespMsg(scaleId int64, data []byte) (*ScaleRespMsg, error) {
	// checkHead, get msgid, get msgtype, check if return code is 0x06, for success
	var err error
	respMsg := &ScaleRespMsg{MsgType: GlastWantRespMsgType, MsgBody: "", ScaleId: scaleId}
	if data[4] == 0x06 {
		respMsg.MsgBody = "ok"
		err = nil
	} else {
		respMsg.MsgBody = "fail"
		err = fmt.Errorf("fail")
	}

	return respMsg, err
}

// read messages from the scale and parses them into data record or command response
// The application runs read in a per-scale goroutine. The application
// ensures that there is at most one reader on a scale by executing all
// reads from this goroutine.
// func (c *Scale) read() {
// 	for {
// 		// read data from scale the pack the msg into a struct then send it to the sendMsg channel
// 		err := readScale(c)
// 		if err == nil && c.queue.capacity != 0 && c.queue.DataLen() > 0 {
// 			msgPacked, err := packMsg(c.queue, true)
// 			if err != nil || msgPacked == nil {
// 				continue
// 			}
// 			msgPacked.ScaleId = c.Id
// 			for _, r := range c.respChans[msgPacked.MsgType] {
// 				r <- *msgPacked
// 			}
// 			//c.respChans[packMsg.MsgType]
// 			//c.srvMgr.recvScaleMsg <- msgPacked
// 		}
// 		time.Sleep(1 * time.Millisecond) // to avoid consume too much cpu time
// 	}
// }

func writeScale(c *Scale, data []byte) error {
	if c.MySerial == nil {
		return fmt.Errorf("scale without a ComPort")
	}

	if len(c.MySerial.sendCh) > SEND_CH_SIZE {
		return fmt.Errorf("serial sendCh full")
	}
	c.MySerial.sendCh <- data
	return nil
}

// write send messages from the hub to scale.
// A goroutine running write is started for each scale. The
// application ensures that there is at most one writer to a scale by
// executing all writes from this goroutine.
func (c *Scale) write() {
	for {
		select {
		case message, ok := <-c.toScaleMsgCh:
			if !ok {
				// The hub closed the channel.
				return
			}
			// send message to scale
			c.MySerial.Write([]byte(message))
		}
	}
}

func (c *Scale) RegisterNotif(msgType RespMsgType, inCh chan *ScaleRespMsg) {
	c.respChans[msgType] = append(c.respChans[msgType], inCh)
}

func (c *Scale) UnRegisterNotif(msgType RespMsgType, inCh chan *ScaleRespMsg) {
	c.respChans[msgType] = remove(c.respChans[msgType], inCh)
}

func remove(s []chan *ScaleRespMsg, m chan *ScaleRespMsg) []chan *ScaleRespMsg {
	for i, msgCh := range s {
		if msgCh == m {
			s[i] = s[len(s)-1]
			return s[:len(s)-1]
		}
	}
	return s
}

func pickerFnOldScale(inData []byte, dataLen int) (packOffset uint, packLen uint, shouldRemoveLen uint) {
	lfcrPos := findLfCrPos(inData, 0, dataLen)
	if lfcrPos > 0 {
		return 0, uint(lfcrPos), uint(lfcrPos) + 2
	} else {
		return 0, 0, 0
	}
}

var RESP_OK_DATA = []byte{0x5a, 0xa5, 0x00, 0x01, 0x06, 0x7f, 0xfb, 0xf1, 0x6e, 0xa5, 0x5a}

func pickerFnTmaxScale(inData []byte, dataLen int) (packOffset uint, packLen uint, shouldRemoveLen uint) {
	// TODO: // find a response package from scale resp data
	// find pattern {5a,a5, 00, 01, 06, 7f, fb, f1, 6e, a5, 5a}
	if len(inData) < len(RESP_OK_DATA) {
		return 0, 0, 0
	}

	// find header
	headPos := findHeadPos(inData, 0, dataLen)
	if headPos == -1 {
		return 0, 0, uint(dataLen) // not found packet
	}

	if len(inData) < headPos+len(RESP_OK_DATA) {
		return 0, 0, uint(headPos)
	}

	for i := 0; i <= len(inData)-len(RESP_OK_DATA); i++ {
		if bytes.Equal(inData[i:len(RESP_OK_DATA)+i], RESP_OK_DATA) {
			log.Log.Debug(fmt.Printf("from serial: %v", inData[i:len(RESP_OK_DATA)+i]))
			return uint(i), uint(len(RESP_OK_DATA)), uint(i + len(RESP_OK_DATA))
		}
	}

	return 0, 0, uint(dataLen)
}

func findLfCrPos(buf []byte, offset int, len int) int {
	foundPos := -1

	if len <= 2 { // \r\n
		return -1
	}

	for pos := offset; pos < len-1; pos++ {
		if buf[pos] == 0x0d && buf[pos+1] == 0x0a { // find a correct response
			foundPos = pos
			break
		}
	}

	return foundPos
}

func pickerFnAutoWeight(inData []byte, dataLen int) (packOffset uint, packLen uint, shouldRemoveLen uint) {
	// find and skip header
	headPos := findHeadPos(inData, 0, len(inData))
	if headPos < 0 {
		return 0, 0, 0
	}
	tmpPackLen := inData[headPos+2]
	if int(tmpPackLen) <= dataLen-2 {
		// check if tail and checksum is correct
		if inData[headPos+int(tmpPackLen)] != 0xA5 || inData[headPos+int(tmpPackLen)+1] != 0x5A {
			return 0, 0, uint(dataLen) // packet error
		}

		if inData[headPos+int(tmpPackLen)-4] != GetChkSum(inData[headPos+2 : headPos+int(tmpPackLen)-4])[0] { // -6: checksum(4 bytes) and tail(2 bytes)
			return 0, 0, uint(dataLen) // remove all data in case checksum error
		}

		if tmpPackLen+2 > 18 {
			log.Log.Warn("Error: remove too much data")
		}
		return uint(headPos), uint(tmpPackLen - 7), uint(tmpPackLen) + 2 // not include cmdid(1), checksum(4) and tail(2)
	}

	return 0, 0, 0
}

func findHeadPos(buf []byte, offset int, len int) int {
	foundPos := -1

	for pos := offset; pos < len-1; pos++ {
		if len > 2 && buf[pos] == CMD_HEAD1 && buf[pos+1] == CMD_HEAD2 { // find a correct response
			foundPos = pos
			break
		}
	}

	return foundPos
}

func extractPack(b []byte) ([]byte, bool) {
	// check if tail and checksum is correct
	if b[len(b)-2] != 0xA5 || b[len(b)-1] != 0x5A {
		return nil, false
	}

	if b[len(b)-6] != GetChkSum(b[:len(b)-6])[0] { // -6: checksum(4 bytes) and tail(2 bytes)
		return nil, false
	}

	return b[2:(len(b) - 6)], true // remove header, checksum and tail
}

func GetChkSum(s []byte) []byte {
	sumArr := make([]byte, 4)
	var sum byte = 0x0
	for _, v := range s {
		sum = sum ^ byte(v)
	}

	sumArr[0] = sum
	sumArr[1] = 0x0
	sumArr[2] = 0x0
	sumArr[3] = 0x0

	return sumArr
}

func parseToScaleReq(reqStr string) (SRequest, error) {
	var req SRequest
	if err := json.UnmarshalFromString(reqStr, &req); err != nil {
		log.Log.Error(err)
		return SRequest{}, err
	}

	return req, nil
}

func procToScaleReq(scale *Scale, req SRequest) error {
	var err error
	switch req.Req {
	case SREQ_GET_WEIGHT:
		// if scale.isBusy {
		// 	srvMgr.recvScaleMsg <- &ScaleRespMsg{ScaleId: scaleId, MsgType: WEIGHT_DATA_RESP, MsgBody: fmt.Errorf("scale is busy")}
		// 	return nil
		// }
		ok := scale.ReadWeight()
		if !ok {
			err = fmt.Errorf("Get weight error")
			msg := ScaleRespMsg{ScaleId: scale.Id, MsgType: WEIGHT_DATA_RESP, MsgBody: err.Error()}
			msgStr, _ := json.MarshalToString(msg)
			scale.client.sendCh <- []byte(msgStr)
			// srvMgr.recvScaleMsg <- &ScaleRespMsg{ScaleId: scaleId, MsgType: WEIGHT_DATA_RESP, MsgBody: err.Error()}
			return nil
		}
		return nil
	case SREQ_ZERO:
		if ok := scale.PerfZero(); !ok {
			err = fmt.Errorf("Perform zero error")
		}
	case SREQ_TARE:
		if ok := scale.PerfTare(); !ok {
			err = fmt.Errorf("Perform tare error")
		}
	case SREQ_REG_WEIGHT_DATA:
		_ = scale.RegWeightData()
	case SREQ_UNREG_WEIGHT_DATA:
		_ = scale.UnRegWeightData()
	case SREQ_GET_RECS:
		recs, _ := scale.GetRecs()
		recsStr, _ := json.MarshalToString(recs)
		resp := &ScaleRespMsg{MsgType: GET_RECS_RESP, MsgBody: recsStr, ScaleId: scale.Id}
		result, _ := json.Marshal(resp)
		scale.client.sendCh <- result
	case SREQ_ADD_REC:
		var rec ScaleRec
		if err := json.Unmarshal([]byte(req.ReqData), &rec); err != nil {
			log.Log.Errorf("Unmarshal ScaleRec error: %v", err)
			resp := &ScaleRespMsg{MsgType: GET_RECS_RESP, MsgBody: err.Error(), ScaleId: scale.Id}
			result, _ := json.Marshal(resp)
			scale.client.sendCh <- result
		} else {
			scale.AddRec(rec)
			resp := &ScaleRespMsg{MsgType: ADD_REC_RESP, MsgBody: "", ScaleId: scale.Id}
			result, _ := json.Marshal(resp)
			scale.client.sendCh <- result
		}
	case SREQ_DEL_REC:
		var id uint64
		if id, err = strconv.ParseUint(req.ReqData, 10, 64); err != nil {
			log.Log.Errorf(err.Error())
			resp := &ScaleRespMsg{MsgType: DEL_REC_RESP, MsgBody: err.Error(), ScaleId: scale.Id}
			result, _ := json.Marshal(resp)
			scale.client.sendCh <- result
		} else {
			_ = scale.DelRec(uint(id))
			resp := &ScaleRespMsg{MsgType: DEL_REC_RESP, MsgBody: "", ScaleId: scale.Id}
			result, _ := json.Marshal(resp)
			scale.client.sendCh <- result
		}
	case SREQ_DOWN_PRN_FMT:
		var resp *ScaleRespMsg
		if err := ReqDownPrnFmt(scale, req.ReqData); err != nil {
			resp = &ScaleRespMsg{MsgType: DOWN_PRN_FMT_RESP, MsgBody: err.Error(), ScaleId: scale.Id}
		} else {
			resp = &ScaleRespMsg{MsgType: DOWN_PRN_FMT_RESP, MsgBody: "", ScaleId: scale.Id}
		}
		result, _ := json.Marshal(resp)
		scale.client.sendCh <- result
	default:
		return fmt.Errorf("unsuported request type")
	}
	return fmt.Errorf("not processed")
}

func ReqDownPrnFmt(c *Scale, csvPrnFmt string) error {
	if prnfmt.ParserFmtToFile(csvPrnFmt) {
		// 读取bin文件
		data, err := ioutil.ReadFile("formatBin.bin")
		if err != nil {
			log.Log.Fatal(err)
		}

		//打开工厂模式

		cmd := enFacModeCmd()
		if res, err := perfCmdNwaitResult(c, cmd, EN_FAC_MODE_RESP); err != nil {
			return err
		} else if res.MsgBody != "ok" {
			return fmt.Errorf("enable factory mode fail")
		} else {
			// do nothing
		}

		//擦除原本秤上的打印格式
		addr := uint32(FLASH_ADDR)
		for i := 0; i < 4; i++ {
			cmd := eraseCmd(uint32(addr))
			if res, err := perfCmdNwaitResult(c, cmd, ERASE_FLASH_RESP); err != nil {
				return err
			} else if res.MsgBody != "ok" {
				return fmt.Errorf("erase fail")
			}
			addr += CMD_ERASE_SIZE
		}

		// 计算数据包数量
		packetCount := len(data) / DATA_LENGTH
		if len(data)%DATA_LENGTH != 0 {
			packetCount += 1
		}

		// 遍历所有数据包
		addr = uint32(FLASH_ADDR)
		for i := 0; i < packetCount; i++ {
			// 计算本包数据
			start := i * DATA_LENGTH
			end := start + DATA_LENGTH
			if end > len(data) {
				end = len(data)
			}
			packetData := data[start:end]

			// 构建数据包
			dataPackCmd := buildSendDataPacket(addr, packetData)

			// 发送数据包
			if res, err := perfCmdNwaitResult(c, dataPackCmd, WRITE_DATA_FLASH_RESP); err != nil {
				return err
			} else if res.MsgBody != "ok" {
				return fmt.Errorf("enable factory mode fail")
			}

			// 地址自增
			addr += 0x100
		}
		log.Log.Info("send bin ok")
	}
	return nil
}

func buildSendDataPacket(addr uint32, data []byte) []byte {
	// 构建包头
	packet := make([]byte, FILE_CHUNK_SIZE)
	binary.BigEndian.PutUint16(packet[0:2], PACKET_HEAD)

	// 构建命令ID与命令类型
	packet[2] = CMD_IDENTIFY
	packet[3] = CMD_TYPE

	// 构建地址
	binary.BigEndian.PutUint32(packet[4:8], addr)

	// 构建数据长度
	binary.BigEndian.PutUint16(packet[8:10], uint16(len(data)))

	// 复制数据
	copy(packet[10:], data)

	// 计算与添加校验码
	checksum := utils.Crc32MPEG2(packet[2 : FILE_CHUNK_SIZE-6])
	binary.BigEndian.PutUint32(packet[FILE_CHUNK_SIZE-6:], checksum)

	// 添加包尾
	binary.BigEndian.PutUint16(packet[FILE_CHUNK_SIZE-2:], PACKET_TAIL)

	return packet
}
