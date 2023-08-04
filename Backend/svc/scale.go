package svc

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math/big"
	"sync"
	"time"

	"tmaxsrv/comm"
	l "tmaxsrv/log"
	"tmaxsrv/prnfmt"
	utils "tmaxsrv/utils"
)

var GlastWantRespMsgType RespMsgType = NO_RESP

const RECV_MAX_BUF_LEN = 2048

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
	SCALE_TIME_OUT_S = 3 // 3 seconds
)

type APInfo struct {
	SeqNo       int    `json:"seqno"`
	Ssid        string `json:"ssid"`
	Rssi        int    `json:"rssi"`
	Mac         string `json:"mac"`
	encryptType string `json:"security"`
}

type MsgBuf struct {
	CurIdx int
	Buf    [RECV_MAX_BUF_LEN]byte
}

type TAddress struct {
	Address string `json:"address"`
	Mask    int    `json:"mask"`
	Proto   string `json:"proto"`
	Family  string `json:"family"`
}

type NetworkConfig struct {
	Addresses []TAddress `json:"address"`
	Gateway   string     `json:"gateway"`
	DNS       []string   `json:"dns"`
	MAC       string     `json:"mac"`
}

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
	EnterAt      time.Time
	respChansMap map[RespMsgType][]chan *ScaleRespMsg
	//recvMsgBufsMap   map[RespMsgType][]MsgBuf // should be removed
	bufs          RingBuffers
	isOldC51Scale bool
	// if client needs unsolicited data from scale
	isSendUnolicitedData bool
	// to scale command is issued and waiting response
	isWaintingResp bool
	// client that will communicate with scale
	client *Client
	// this channel is to inform the procScaleRespMessage goroutine, which process the data from serial port, to quit
	quitProcScaleRespMessageCh chan bool
	// this channel is to inform the procToScaleMsg, which process the data from serial port, to quit
	quitProcToScaleMsgCh chan bool
}

// NewScale creates a new scale
func NewScale(scaleMgr *ScaleMgr, conn *ScaleConnMedia, model string, sn string, isTest bool) (*Scale, error) {
	var pcnf ComInfo
	var sport *TSerial
	var err error
	if conn.TMedia == MEDIA_COM {
		// open COM connection
		if err := json.Unmarshal([]byte(conn.MediaConf.MediaInfoJson), &pcnf); err != nil {
			l.Log.Errorf("error unmarshalling: %v", err)
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
			l.Log.Error("not support scale type")
			return nil, fmt.Errorf("not support scale type")
		}
	}
	scale := &Scale{scaleMgr: scaleMgr, Conn: conn, Model: model, Sn: sn, toScaleMsgCh: make(chan string, SCALE_SEND_CH_SIZE),
		fromScaleMsgCh: make(chan string, SCALE_RECV_CH_SIZE), MySerial: sport,
		quitProcScaleRespMessageCh: make(chan bool, 1), quitProcToScaleMsgCh: make(chan bool, 1)}
	scale.respChansMap = map[RespMsgType][]chan *ScaleRespMsg{}

	var respTypes []string
	for _, respType := range cmdsRespMap {
		respTypes = append(respTypes, string(respType))
	}

	// create buffer for each message
	scale.bufs = NewRingBuffers(respTypes...)

	responseChannels := make(map[RespMsgType]chan interface{})

	for _, respType := range cmdsRespMap {
		if _, ok := responseChannels[respType]; !ok {
			responseChannels[respType] = make(chan interface{})
		}
	}

	// example usage: send a message to the WEIGHT_DATA_RESP channel
	// weightDataRespCh := responseChannels[WEIGHT_DATA_RESP]
	// weightDataRespCh <- "some message"

	scale.isOldC51Scale = false // in this moment we use this software for old C51 scales, for T-Max scales should set this to false

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
	l.Log.Warn("Client disconnected, HandleClientDisconnect called")
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
		case <-s.quitProcScaleRespMessageCh:
			quit = true
		case inPack := <-s.MySerial.recvCh:
			if inPack.PayloadLen == 0 { // TODO: maybe caused by closed serial port
				continue
			}
			l.Log.Debugf("From sport: %v", inPack)
			// if !s.isSendUnolicitedData || !s.isWaintingResp {
			// 	// continue // FIXME: skip this line for testing purpose
			// }
			if comm.GcurScale == comm.SCALE_TYPE_OLD_C51 {
				// TODO:
			} else if comm.GcurScale == comm.SCALE_TYPE_TMAX {
				// find message buffer that associate to the message
				var cmdHex uint16 = (uint16(inPack.CmdID) << 8) | uint16(inPack.CmdSubId)
				bufName := cmdsRespMap[CmdID(cmdHex)]
				if bufName == "" {
					l.Log.Errorf("error on getting bufName for the cmd: %v", cmdHex)
					continue
				}
				fmt.Println(bufName)
				s.bufs.Write(string(bufName), inPack.Payload)
				msg := extractMessage(s.Id, s.bufs[string(bufName)], bufName)
				if (msg != ScaleRespMsg{}) {
					chs := s.respChansMap[msg.MsgType]
					for _, ch := range chs {
						if len(ch) == 0 { // to avoid blocking, this kind of channel should only be used once
							ch <- &msg
						}
					}
					if len(s.fromScaleMsgCh) < RECV_CH_SIZE { // no use in this moment
						if msg.MsgType == WEIGHT_DATA && !s.isSendUnolicitedData { // skip sending weight data to client if it doesn't not register this message
							continue
						}
						if msgStr, err := json.MarshalToString(msg); err == nil {
							if s.client != nil {
								fmt.Println("%%%%%%%%%%%%%%: " + msgStr)
								s.client.sendCh <- []byte(msgStr)
							}
							//s.fromScaleMsgCh <- msgStr
						} else {
							l.Log.Errorf("marshal msg err: %v", err.Error())
						}
					} else {
						l.Log.Error("fromScaleMsgCh full")
					}
				} else {
					l.Log.Errorf("extractMsgTmaxScale error")
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
		case <-s.quitProcToScaleMsgCh:
			quit = true
		case userMessage, ok := <-s.client.recvCh:
			if !ok {
				l.Log.Error("client's recvCh closed")
				continue
			}
			var data map[string][]byte
			json.Unmarshal(userMessage, &data)
			l.Log.Debugf("userMessage: %v", data)
			scaleId := new(big.Int).SetBytes(data["scaleId"]).Int64()
			if scaleId != s.Id {
				l.Log.Errorf("wrong scale id :%v received, our Id is: %v", scaleId, s.Id)
				continue
			}
			if data == nil { // TODO: maybe caused by ...client?
				continue
			}
			l.Log.Debugf("From wsclient: %v", string(data["message"]))
			if req, err := parseToScaleReq(string(data["message"])); err == nil {
				go procToScaleReq(s, req) // TODO: handle error
			}
		}
	}
}

func (s *Scale) ModifyMedia(conf MediaConf) bool {
	//	c.Lock()
	//	defer c.Unlock()
	if conf.Type == MEDIA_COM {
		var pcnf ComInfo
		var err error
		// open COM connection
		if err := json.Unmarshal([]byte(conf.MediaInfoJson), &pcnf); err != nil {
			l.Log.Errorf("error unmarshalling: %v", err)
		}
		var pickFun packPickerFn = nil
		if s.MySerial != nil {
			s.quitProcScaleRespMessageCh <- true
			s.quitProcToScaleMsgCh <- true
			pickFun = s.MySerial.pickerFn
			s.MySerial.Close()
			s.MySerial = nil
			if s.MySerial, err = NewSerial(pcnf, pickFun); err != nil {
				l.Log.Error(err.Error())
			}
		} else {
			l.Log.Error("no serial port is assigned before")
			return false
		}

	} else if conf.Type == MEDIA_NET {
		return false

	} else if conf.Type == MEDIA_BT {
		return false

	} else {

	}
	go s.procScaleRespMsg()
	go s.procToScaleMsg()
	return true
}

var mu sync.Mutex

func (c *Scale) RegisterNotif(msgType RespMsgType, inCh chan *ScaleRespMsg) {
	addNotif(c, msgType, inCh)
}

func (c *Scale) UnRegisterNotif(msgType RespMsgType, inCh chan *ScaleRespMsg) {
	removeNotif(c, msgType, inCh)
}

func addNotif(s *Scale, msgType RespMsgType, inCh chan *ScaleRespMsg) {
	mu.Lock()
	defer mu.Unlock()
	s.respChansMap[msgType] = append(s.respChansMap[msgType], inCh)
}

func removeNotif(s *Scale, msgType RespMsgType, inCh chan *ScaleRespMsg) {
	mu.Lock()
	defer mu.Unlock()
	s.respChansMap[msgType] = remove(s.respChansMap[msgType], inCh)
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

func parseToScaleReq(reqStr string) (SRequest, error) {
	var req SRequest
	if err := json.UnmarshalFromString(reqStr, &req); err != nil {
		l.Log.Error(err)
		return SRequest{}, err
	}

	return req, nil
}

func ReqModifyBTName(s *Scale, name string) error {
	if err := EnFacMode(s); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	if err := EnPassthrough(s); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	if err := ModifyBTName(s, name); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	if err := DisPassthrough(s); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	return nil
}

func ReqSendDataToBT(s *Scale, data string) error {
	if err := EnFacMode(s); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	if err := EnPassthrough(s); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	if err := SendDataToBT(s, data); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	if err := DisPassthrough(s); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	return nil
}

func ReqSendDataToWifi(s *Scale, data string) error {
	if err := EnFacMode(s); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	if err := EnPassthrough(s); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	if err := SendDataToWifi(s, data); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	if err := DisPassthrough(s); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	return nil
}

func ReqGetApList(s *Scale) error {

	if err := EnFacMode(s); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	if err := EnPassthrough(s); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	if err := GetApList(s); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	if err := DisPassthrough(s); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	return nil
}

func ReqConnectAp(s *Scale, ssid string, password string, bssid string) error {

	if err := EnFacMode(s); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	if err := EnPassthrough(s); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	if err := ConnectWifiAp(s, ssid, password, bssid); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	if err := DisPassthrough(s); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	return nil
}

func ReqSetWifiStaticIp(s *Scale, ip string, gateway string, netmask string) error {
	if err := EnFacMode(s); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	if err := EnPassthrough(s); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	if err := SetWifiStaticIp(s, ip, gateway, netmask); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	if err := DisPassthrough(s); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	return nil
}

func ReqGetIpInfo(s *Scale) error {

	if err := EnFacMode(s); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	if err := EnPassthrough(s); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	if err := GetIpInfo(s); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	if err := DisPassthrough(s); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	return nil
}

func ReqDownPrnFmt(c *Scale, csvPrnFmt string) error {
	if !gIsKeyValid {
		return fmt.Errorf("license key is not valid")
	}

	layout := "2006-01-02"
	date, err := time.Parse(layout, gLicValidDate)
	if err != nil || time.Now().After(date) {
		fmt.Println(err)
		return fmt.Errorf("license expired")
	}

	if prnfmt.ParserFmtToFile(csvPrnFmt) {
		// 读取bin文件
		data, err := ioutil.ReadFile("formatBin.bin")
		if err != nil {
			l.Log.Fatal(err)
		}

		//打开工厂模式

		l.Log.Debug("send enable factory mode cmd to scale")
		cmd := EN_ENG_CMD
		if res, err := perfCmdNwaitResult(c, cmd, EN_FAC_MODE_RESP); err != nil {
			return err
		} else if res.MsgBody != "ok" {
			return fmt.Errorf("enable factory mode fail")
		} else {
			// do nothing
		}

		//擦除原本秤上的打印格式
		l.Log.Debug("erase flash on scale")
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
		l.Log.Debug("send data package to scale")
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
		l.Log.Info("send bin ok")
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
