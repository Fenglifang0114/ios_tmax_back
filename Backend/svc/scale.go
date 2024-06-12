package svc

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"math/big"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"tmaxsrv/cmd"
	m "tmaxsrv/comm"
	"tmaxsrv/eeprom"
	l "tmaxsrv/log"
	"tmaxsrv/picker"
	"tmaxsrv/prnfmt"
	utils "tmaxsrv/util"
)

const (
	// 	FLASH_ADDR_TMAX                = 0x08003000           //FLF
	// 	PACKET_HEAD_TMAX               = 0x5AA5
	// 	CMD_IDENTIFY_TMAX              = 0xA0
	// 	CMD_TYPE_TMAX                  = 0xB0
	// 	PACKET_TAIL_TMAX               = 0xA55A
	DATA_LENGTH_256_TMAX = 256
	DATA_LENGTH_512_TMAX = 512

// FILE_CHUNK_SIZE_TMAX           = 272
// OPEN_FAC_CHUNK_SIZE_TMAX       = 6
// CLOSE_FAC_CHUNK_SIZE_TMAX      = 6
// EN_PASSTH_CHUNK_SIZE_TMAX      = 6
// DIS_PASSTH_CHUNK_SIZE_TMAX     = 6
// MODIFY_BT_NAME_CHUNK_SIZE_TMAX = 24
// REC_CHUNK_SIZE_TMAX            = 11
// EARSE_CHUNK_SIZE_TMAX          = 0x10
// CMD_ERASE_TMAX                 = 0xA2
// CMD_ERASE_SIZE_TMAX            = 0x800
// CMD_FLASH_TMAX                 = 0xB0
)

var GlastWantRespMsgType m.RespMsgType = m.NO_RESP

const RECV_MAX_BUF_LEN = 2048

const MD5SEED = "*T-Scale*"

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

const (
	WIFI_BT_INFO            = "wifi_or_bt"
	WIFI_BT_INFO_FIELD_NAME = "M_WIRELESS_PERIPHERAL_TYPE"
)

type APInfo struct {
	SeqNo       int    `json:"seqno"`
	Ssid        string `json:"ssid"`
	Rssi        int    `json:"rssi"`
	Mac         string `json:"mac"`
	encryptType string `json:"security"`
}

type WifiInfo struct {
	Ssid  string `json:"ssid"`
	Rssi  int    `json:"rssi"`
	Bssid string `json:"mac"`
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
	// MediaConf string
	// com port
	MySerial *TSerial
	// Network socket
	// tcpSocket net.Socket
	// Bluetooth
	// btConn BtCom
	// scale Id
	Id int64
	// scale Model, if not supported then the default Model is "legacy"
	Model string
	// scale serial number, if not supported then the default serial number is "123456789"
	ScaleCat m.ScaleCat // scale type: C51, T2200, JWP, TMAX
	Sn       string
	// send to scale channel, message will be json string
	toScaleMsgCh chan string
	// receive from scale channel, message will be json string
	fromScaleMsgCh chan string
	// added time
	EnterAt      time.Time
	respChansMap map[m.RespMsgType][]chan *ScaleRespMsg
	// recvMsgBufsMap   map[RespMsgType][]MsgBuf // should be removed
	bufs utils.RingBuffers
	//isOldC51Scale bool
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

	// function pointer to handle message from scale
	// composer object
	composer *m.CmdComposer
	Pcnf     ComInfo

	isScalePassth    bool
	IsScalePassthHex bool
}

// NewScale creates a new scale
func NewScale(scaleMgr *ScaleMgr, conn *ScaleConnMedia, scaleCat m.ScaleCat, model string, sn string, isTest bool) (*Scale, error) {
	var pcnf ComInfo
	var sport *TSerial
	if conn.TMedia == MEDIA_COM {
		// open COM connection
		if err := json.Unmarshal([]byte(conn.MediaConf.MediaInfoJson), &pcnf); err != nil {
			l.Log.Errorf("error unmarshalling: %v", err)
		}

		picker := picker.GetPickerFn(scaleCat)
		sport, _ = NewSerial(pcnf, picker)
	}
	scale := &Scale{
		scaleMgr: scaleMgr, Conn: conn, ScaleCat: scaleCat, Model: model, Sn: sn, toScaleMsgCh: make(chan string, SCALE_SEND_CH_SIZE),
		fromScaleMsgCh: make(chan string, SCALE_RECV_CH_SIZE), MySerial: sport, Pcnf: pcnf,
		quitProcScaleRespMessageCh: make(chan bool, 1), quitProcToScaleMsgCh: make(chan bool, 1),
	}
	scale.respChansMap = map[m.RespMsgType][]chan *ScaleRespMsg{}
	composer := cmdComposerFuncMap[scale.ScaleCat]
	scale.composer = &composer

	var respTypes []string
	for _, respType := range utils.CmdsRespMap {
		respTypes = append(respTypes, string(respType))
	}

	// create buffer for each message
	scale.bufs = utils.NewRingBuffers(respTypes...)

	responseChannels := make(map[m.RespMsgType]chan interface{})

	for _, respType := range utils.CmdsRespMap {
		if _, ok := responseChannels[respType]; !ok {
			responseChannels[respType] = make(chan interface{})
		}
	}

	// example usage: send a message to the WEIGHT_DATA_RESP channel
	// weightDataRespCh := responseChannels[WEIGHT_DATA_RESP]
	// weightDataRespCh <- "some message"

	scale.isSendUnolicitedData = false
	scale.isWaintingResp = false
	scale.isScalePassth = false
	scale.IsScalePassthHex = false

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
			if inPack.PayloadLen == 0 {
				continue
			}
			l.Log.Debugf("From sport: %v", inPack)

			if s.ScaleCat == m.SCALE_C51 {
			} else if s.ScaleCat == m.SCALE_T2200 {
				msg, err := retreiveRespMsgT2200(s.Id, inPack.Payload)
				if err != nil {
					continue
				}
				sendMsgIntoChsOrWeightToClient(s, msg)
			} else if s.ScaleCat == m.SCALE_TMAX || s.ScaleCat == m.SCALE_TMAX_PASSTH {
				// find message buffer that associate to the message

				if s.isScalePassth {
					var cmdHex uint16 = cmd.CMDID_SCALE_PASSTH_DATA_TMAX
					bufName := utils.CmdsRespMap[utils.CmdID(cmdHex)]
					if bufName == "" {
						l.Log.Errorf("error on getting bufName for the cmd: %v", cmdHex)
						continue
					}

					fmt.Println(bufName)
					s.bufs.Write(string(bufName), inPack.Payload)
					msg := extractScalePassthDataTMAX(s, s.bufs[string(bufName)], bufName)
					sendMsgIntoChsOrWeightToClient(s, &msg) //20231023  @111
					continue
				}
				var cmdHex uint16 = (uint16(inPack.CmdID) << 8) | uint16(inPack.CmdSubId)
				bufName := utils.CmdsRespMap[utils.CmdID(cmdHex)]
				if bufName == "" {
					l.Log.Errorf("error on getting bufName for the cmd: %v", cmdHex)
					continue
				}
				fmt.Println(bufName)
				s.bufs.Write(string(bufName), inPack.Payload)
				msg := extractMessageTMAX(s.Id, s.bufs[string(bufName)], bufName)
				sendMsgIntoChsOrWeightToClient(s, &msg) //20231023  @111
			} else {

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
	mu.Lock()
	defer mu.Unlock()
	if conf.Type == MEDIA_COM {
		var pcnf ComInfo
		var err error
		// open COM connection
		if err := json.Unmarshal([]byte(conf.MediaInfoJson), &pcnf); err != nil {
			l.Log.Errorf("error unmarshalling: %v", err)
		}
		var pickFun picker.PickerFunc = nil
		if s.MySerial != nil {
			s.quitProcScaleRespMessageCh <- true
			s.quitProcToScaleMsgCh <- true
			pickFun = s.MySerial.pickerFn
			s.MySerial.Close()
			s.MySerial = nil
		} else {
			l.Log.Error("no serial port is assigned before")
			return false
		}
		time.Sleep(1 * time.Second)
		if s.MySerial, err = NewSerial(pcnf, pickFun); err != nil {
			l.Log.Error(err.Error())
		}
		s.Pcnf = pcnf
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

func (c *Scale) RegisterNotif(msgType m.RespMsgType, inCh chan *ScaleRespMsg) {
	addNotif(c, msgType, inCh)
}

func (c *Scale) UnRegisterNotif(msgType m.RespMsgType, inCh chan *ScaleRespMsg) {
	removeNotif(c, msgType, inCh)
}

func addNotif(s *Scale, msgType m.RespMsgType, inCh chan *ScaleRespMsg) {
	mu.Lock()
	defer mu.Unlock()
	s.respChansMap[msgType] = append(s.respChansMap[msgType], inCh)
}

func removeNotif(s *Scale, msgType m.RespMsgType, inCh chan *ScaleRespMsg) {
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

func enablePassthrough(s *Scale, respType m.RespMsgType) (*ScaleRespMsg, error) {
	// msg, err := EnFacMode(s)
	// if err != nil {
	// 	return &ScaleRespMsg{respType, "fail", s.Id}, err
	// }
	// if msg.MsgBody != "ok" {
	// 	return &ScaleRespMsg{respType, "fail", s.Id}, nil
	// }
	// time.Sleep(100 * time.Millisecond)
	if _, err := EnPassthrough(s); err != nil {
		return &ScaleRespMsg{respType, "fail", s.Id}, err
	}
	time.Sleep(100 * time.Millisecond)
	return &ScaleRespMsg{respType, "ok", s.Id}, nil
}

func ReqModifyBTName(s *Scale, name string) (*ScaleRespMsg, error) {
	defer DisPassthrough(s)
	msg, err := enablePassthrough(s, m.MODIFY_BT_NAME_RESP)
	if msg.MsgBody != "ok" {
		return msg, err
	}
	return s.ModifyBTName(name)
}

func ReqSendDataToBT(s *Scale, data string) (*ScaleRespMsg, error) {
	defer DisPassthrough(s)
	msg, err := enablePassthrough(s, m.SEND_DATA_TO_BT_RESP)
	if msg.MsgBody != "ok" {
		return msg, err
	}

	return SendDataToBT(s, data+"\r\n\x00")
}

func ReqSendDataToWifi(s *Scale, data string) (*ScaleRespMsg, error) {
	defer DisPassthrough(s)
	msg, err := enablePassthrough(s, m.SEND_DATA_TO_WIFI_RESP)
	if msg.MsgBody != "ok" {
		return msg, err
	}

	return SendDataToWifi(s, data)
}

func ReqGetWifiApInfo(s *Scale) (*ScaleRespMsg, error) {
	defer DisPassthrough(s)
	msg, err := enablePassthrough(s, m.GET_WIFI_AP_INFO_RESP)
	if msg.MsgBody != "ok" {
		return msg, err
	}

	return GetWifiApInfo(s)
}

func ReqGetApList(s *Scale) (*ScaleRespMsg, error) {
	defer DisPassthrough(s)
	msg, err := enablePassthrough(s, m.GET_AP_LIST_RESP)
	if msg.MsgBody != "ok" {
		return msg, err
	}

	return GetApList(s)
}

func ReqConnectAp(s *Scale, ssid string, password string, bssid string) (*ScaleRespMsg, error) {
	//defer DisPassthrough(s)
	msg, err := enablePassthrough(s, m.CONNECT_AP_RESP)
	if msg.MsgBody != "ok" {
		return msg, err
	}

	res, err := ConnectWifiAp(s, ssid, password, bssid)
	return res, err
}

func ReqConnectApOneKey(s *Scale, ssid string, password string, bssid string) (*ScaleRespMsg, error) {
	//defer DisPassthrough(s)
	msg, err := enablePassthrough(s, m.CONNECT_AP_RESP)
	if msg.MsgBody != "ok" {
		return msg, err
	}

	res, err := ConnectWifiApOneKey(s, ssid, password, bssid)
	return res, err
}

func ReqSetWifiDynamicIp(s *Scale) (*ScaleRespMsg, error) {
	defer DisPassthrough(s)
	msg, err := enablePassthrough(s, m.SET_WIFI_STATIC_IP_RESP)
	if msg.MsgBody != "ok" {
		return msg, err
	}

	return SetWifiDynamicIp(s)
}

func ReqSetWifiStaticIp(s *Scale, ip string, gateway string, netmask string) (*ScaleRespMsg, error) {
	defer DisPassthrough(s)
	msg, err := enablePassthrough(s, m.SET_WIFI_STATIC_IP_RESP)
	if msg.MsgBody != "ok" {
		return msg, err
	}

	return SetWifiStaticIp(s, ip, gateway, netmask)
}

func ReqGetIpInfo(s *Scale) (*ScaleRespMsg, error) {
	defer DisPassthrough(s)
	msg, err := enablePassthrough(s, m.SET_WIFI_STATIC_IP_RESP)
	if msg.MsgBody != "ok" {
		return msg, err
	}

	return GetIpInfo(s)
}

func ReqChangeWifiMode(s *Scale, req SRequest) (*ScaleRespMsg, error) {
	defer DisPassthrough(s)
	msg, err := enablePassthrough(s, m.CHANGE_WIFI_MODE_RESP)
	if msg.MsgBody != "ok" {
		return msg, err
	}
	return ChangeWifiMode(s)
}

func ReqGetIpMode(s *Scale) (*ScaleRespMsg, error) {
	defer DisPassthrough(s)
	msg, err := enablePassthrough(s, m.GET_IP_MODE_RESP)
	if msg.MsgBody != "ok" {
		return msg, err
	}

	return GetIpMode(s)
}

func ReqModifyEepromInfo(s *Scale, req SRequest) (*ScaleRespMsg, error) {
	var eepromData EepromDataResp
	if err := json.UnmarshalFromString(req.ReqData, &eepromData.EepromData); err != nil {
		return &ScaleRespMsg{}, err
	}
	modifyData := eepromData.EepromData

	print(len(modifyData))
	composer := s.composer
	fn := composer.ComposeCmd
	for i := 0; i < len(modifyData); i++ {
		if modifyData[i].Type == "string" {
			packetData := stringToLittleEndianBytes(modifyData[i].CurrValue, modifyData[i].Size)
			packDataHexStr := hex.EncodeToString(packetData)
			cmd, timeoutMs, err := fn(composer, m.CMD_WRITE_EEPROM, m.CmdData{Type: m.DATA_TYPE_STR, Data: fmt.Sprintf("%08x:%s", modifyData[i].Addr, packDataHexStr)})
			if err != nil {
				return &ScaleRespMsg{}, err
			}

			// 发送数据包
			if res, err := perfCmdNwaitResult(s, cmd, m.WRITE_DATA_FLASH_RESP, timeoutMs); err != nil {
				return &ScaleRespMsg{}, err
			} else if res.MsgBody != "ok" {
				return &ScaleRespMsg{}, fmt.Errorf("write eeprom fail")
			}
		} else if modifyData[i].Type == "int" {
			if modifyData[i].Size == 1 {
				num, err := strconv.Atoi(modifyData[i].CurrValue)
				if err != nil {

					return &ScaleRespMsg{m.MODIFY_EEPROM_INFO_RESP, "fail,data error", s.Id}, nil
				}
				byteVal := []byte{byte(num)}
				packDataHexStr := hex.EncodeToString(byteVal)
				cmd, timeoutMs, err := fn(composer, m.CMD_WRITE_EEPROM, m.CmdData{Type: m.DATA_TYPE_STR, Data: fmt.Sprintf("%08x:%s", modifyData[i].Addr, packDataHexStr)})
				if err != nil {
					return &ScaleRespMsg{}, err
				}
				// 发送数据包
				if res, err := perfCmdNwaitResult(s, cmd, m.WRITE_DATA_FLASH_RESP, timeoutMs); err != nil {
					return &ScaleRespMsg{}, err
				} else if res.MsgBody != "ok" {
					return &ScaleRespMsg{}, fmt.Errorf("write eeprom fail")
				}
			} else if modifyData[i].Size == 4 && modifyData[i].SubType == "ip" {
				packetData := ipv4StringToBytes(modifyData[i].CurrValue)
				packDataHexStr := hex.EncodeToString(packetData)
				cmd, timeoutMs, err := fn(composer, m.CMD_WRITE_EEPROM, m.CmdData{Type: m.DATA_TYPE_STR, Data: fmt.Sprintf("%08x:%s", modifyData[i].Addr, packDataHexStr)})
				if err != nil {
					return &ScaleRespMsg{}, err
				}
				// 发送数据包
				if res, err := perfCmdNwaitResult(s, cmd, m.WRITE_DATA_FLASH_RESP, timeoutMs); err != nil {
					return &ScaleRespMsg{}, err
				} else if res.MsgBody != "ok" {
					return &ScaleRespMsg{}, fmt.Errorf("write eeprom fail")
				}
			} else if modifyData[i].Size == 2 || modifyData[i].Size == 4 || modifyData[i].Size == 8 {
				packetData, _ := intToLittleEndianBytes(modifyData[i].CurrValue, modifyData[i].Size)
				packDataHexStr := hex.EncodeToString(packetData)
				cmd, timeoutMs, err := fn(composer, m.CMD_WRITE_EEPROM, m.CmdData{Type: m.DATA_TYPE_STR, Data: fmt.Sprintf("%08x:%s", modifyData[i].Addr, packDataHexStr)})
				if err != nil {
					return &ScaleRespMsg{}, err
				}
				// 发送数据包
				if res, err := perfCmdNwaitResult(s, cmd, m.WRITE_DATA_FLASH_RESP, timeoutMs); err != nil {
					return &ScaleRespMsg{}, err
				} else if res.MsgBody != "ok" {
					return &ScaleRespMsg{}, fmt.Errorf("write eeprom fail")
				}
			}

		} else if modifyData[i].Type == "double" {
			if modifyData[i].Size == 8 {
				packetData := StringToFloat64Bytes(modifyData[i].CurrValue)
				packDataHexStr := hex.EncodeToString(packetData)
				cmd, timeoutMs, err := fn(composer, m.CMD_WRITE_EEPROM, m.CmdData{Type: m.DATA_TYPE_STR, Data: fmt.Sprintf("%08x:%s", modifyData[i].Addr, packDataHexStr)})
				if err != nil {
					return &ScaleRespMsg{}, err
				}
				// 发送数据包
				if res, err := perfCmdNwaitResult(s, cmd, m.WRITE_DATA_FLASH_RESP, timeoutMs); err != nil {
					return &ScaleRespMsg{}, err
				} else if res.MsgBody != "ok" {
					return &ScaleRespMsg{}, fmt.Errorf("write eeprom fail")
				}
			}
		}
	}
	return &ScaleRespMsg{m.MODIFY_EEPROM_INFO_RESP, "ok", s.Id}, nil
}

type ServerIpData struct {
	Ip         string `json:"ip"`
	Netmask    string `json:"netmask"`
	Gateway    string `json:"gateway"`
	ServerIp   string `json:"serverIp"`
	ServerPort string `json:"serverPort"`
}

type ServerIpAddr struct {
	IpAddr         int
	NetmaskAddr    int
	GatewayAddr    int
	ServerIpAddr   int
	ServerPortAddr int
}

func getServerAddr() ServerIpAddr {
	var serverAddr ServerIpAddr
	var eepromResp EepromDataResp
	fieldData, res := eeprom.GetExcelData()
	if !res {
		return serverAddr
	}
	filedDataLen := len(fieldData.EepromStruct)
	if filedDataLen <= 0 {
		return serverAddr
	}
	for i := 0; i < filedDataLen; i++ {
		var eData EData
		eData.FiledName = fieldData.EepromStruct[i].FieldName
		eData.Permission = fieldData.EepromStruct[i].Permission
		eData.Comment = fieldData.EepromStruct[i].Comments
		eData.Category = fieldData.EepromStruct[i].Category
		eData.Addr = fieldData.EepromStruct[i].Addr
		eData.Size = fieldData.EepromStruct[i].Size
		eData.Description = fieldData.EepromStruct[i].Description
		eData.Values = fieldData.EepromStruct[i].Values
		eData.CurrValue = ""
		eData.Type = fieldData.EepromStruct[i].Type
		eData.SubType = fieldData.EepromStruct[i].SubType
		eepromResp.EepromData = append(eepromResp.EepromData, eData)
	}

	count := 0
	for _, eData := range eepromResp.EepromData {
		if eData.FiledName == "M_SCALE_IP" || eData.FiledName == "M_SCALE_MASK" || eData.FiledName == "M_SCALE_GATEWAY" || eData.FiledName == "M_PORT_NUMBER" || eData.FiledName == "M_NET_IP_ADDRESS" {
			fmt.Printf("FieldName: %s, Addr: %d\n", eData.FiledName, eData.Addr)
			if eData.FiledName == "M_SCALE_IP" {
				serverAddr.IpAddr = eData.Addr
			}
			if eData.FiledName == "M_SCALE_MASK" {
				serverAddr.NetmaskAddr = eData.Addr
			}
			if eData.FiledName == "M_SCALE_GATEWAY" {
				serverAddr.GatewayAddr = eData.Addr
			}
			if eData.FiledName == "M_NET_IP_ADDRESS" {
				serverAddr.ServerIpAddr = eData.Addr
			}
			if eData.FiledName == "M_PORT_NUMBER" {
				serverAddr.ServerPortAddr = eData.Addr
			}
			count++
			if count == 6 {
				break // 找到五个后跳出循环
			}
		}
	}
	return serverAddr

}

//ReqEnFactoryMode TODO:

func ReqSetServerIp(s *Scale, req SRequest) (*ScaleRespMsg, error) {
	var serverData ServerIpData
	if err := json.UnmarshalFromString(req.ReqData, &serverData); err != nil {
		return &ScaleRespMsg{}, err
	}
	reg, err, res := openFactory(s)
	if err != nil || !res {
		return reg, err
	}

	serverAddr := getServerAddr()
	if serverAddr.IpAddr < 1 && serverAddr.GatewayAddr < 1 && serverAddr.NetmaskAddr < 1 && serverAddr.ServerIpAddr < 1 && serverAddr.ServerPortAddr < 1 {
		return &ScaleRespMsg{}, fmt.Errorf("get eeprom addr fail")
	}

	composer := s.composer
	fn := composer.ComposeCmd

	if serverData.Ip != "" {
		packetData := ipv4StringToBytes(serverData.Ip)
		packDataHexStr := hex.EncodeToString(packetData)
		cmd, timeoutMs, err := fn(composer, m.CMD_WRITE_EEPROM, m.CmdData{Type: m.DATA_TYPE_STR, Data: fmt.Sprintf("%08x:%s", serverAddr.IpAddr, packDataHexStr)})
		if err != nil {
			return &ScaleRespMsg{}, err
		}
		// 发送数据包
		if res, err := perfCmdNwaitResult(s, cmd, m.WRITE_DATA_FLASH_RESP, timeoutMs); err != nil {
			return &ScaleRespMsg{}, err
		} else if res.MsgBody != "ok" {
			return &ScaleRespMsg{}, fmt.Errorf("write eeprom fail")
		}
	}
	if serverData.Gateway != "" {
		packetData := ipv4StringToBytes(serverData.Gateway)
		packDataHexStr := hex.EncodeToString(packetData)
		cmd, timeoutMs, err := fn(composer, m.CMD_WRITE_EEPROM, m.CmdData{Type: m.DATA_TYPE_STR, Data: fmt.Sprintf("%08x:%s", serverAddr.GatewayAddr, packDataHexStr)})
		if err != nil {
			return &ScaleRespMsg{}, err
		}
		// 发送数据包
		if res, err := perfCmdNwaitResult(s, cmd, m.WRITE_DATA_FLASH_RESP, timeoutMs); err != nil {
			return &ScaleRespMsg{}, err
		} else if res.MsgBody != "ok" {
			return &ScaleRespMsg{}, fmt.Errorf("write eeprom fail")
		}
	}

	if serverData.Netmask != "" {
		packetData := ipv4StringToBytes(serverData.Netmask)
		packDataHexStr := hex.EncodeToString(packetData)
		cmd, timeoutMs, err := fn(composer, m.CMD_WRITE_EEPROM, m.CmdData{Type: m.DATA_TYPE_STR, Data: fmt.Sprintf("%08x:%s", serverAddr.NetmaskAddr, packDataHexStr)})
		if err != nil {
			return &ScaleRespMsg{}, err
		}
		// 发送数据包
		if res, err := perfCmdNwaitResult(s, cmd, m.WRITE_DATA_FLASH_RESP, timeoutMs); err != nil {
			return &ScaleRespMsg{}, err
		} else if res.MsgBody != "ok" {
			return &ScaleRespMsg{}, fmt.Errorf("write eeprom fail")
		}
	}

	if serverData.ServerIp != "" {
		packetData := ipv4StringToBytes(serverData.ServerIp)
		packDataHexStr := hex.EncodeToString(packetData)
		cmd, timeoutMs, err := fn(composer, m.CMD_WRITE_EEPROM, m.CmdData{Type: m.DATA_TYPE_STR, Data: fmt.Sprintf("%08x:%s", serverAddr.ServerIpAddr, packDataHexStr)})
		if err != nil {
			return &ScaleRespMsg{}, err
		}
		// 发送数据包
		if res, err := perfCmdNwaitResult(s, cmd, m.WRITE_DATA_FLASH_RESP, timeoutMs); err != nil {
			return &ScaleRespMsg{}, err
		} else if res.MsgBody != "ok" {
			return &ScaleRespMsg{}, fmt.Errorf("write eeprom fail")
		}
	}

	if serverData.ServerPort != "" {
		packetData, _ := intToLittleEndianBytes(serverData.ServerPort, 2)
		packDataHexStr := hex.EncodeToString(packetData)
		cmd, timeoutMs, err := fn(composer, m.CMD_WRITE_EEPROM, m.CmdData{Type: m.DATA_TYPE_STR, Data: fmt.Sprintf("%08x:%s", serverAddr.ServerPortAddr, packDataHexStr)})
		if err != nil {
			return &ScaleRespMsg{}, err
		}
		// 发送数据包
		if res, err := perfCmdNwaitResult(s, cmd, m.WRITE_DATA_FLASH_RESP, timeoutMs); err != nil {
			return &ScaleRespMsg{}, err
		} else if res.MsgBody != "ok" {
			return &ScaleRespMsg{}, fmt.Errorf("write eeprom fail")
		}
	}

	return &ScaleRespMsg{m.SET_SERVER_IP_RESP, "ok", s.Id}, nil

}

type HeaderData struct {
	ID    int    `json:"id"`
	Value string `json:"value"`
}

type HeaderList struct {
	ListData []HeaderData `json:"listData"`
}

func ReqModifyVarValue(s *Scale, req SRequest) (*ScaleRespMsg, error) {
	reg, err, res := openFactory(s)
	if err != nil || !res {
		return reg, err
	}
	var headerList HeaderList
	err = json.Unmarshal([]byte(req.ReqData), &headerList)
	if err != nil {
		return &ScaleRespMsg{}, err
	}

	composer := s.composer
	fn := composer.ComposeCmd
	var sendArray []byte
	for i := 0; i < len(headerList.ListData); i++ {
		tmpData := headerList.ListData[i]
		idByte := make([]byte, 2)
		dataLen := make([]byte, 2)
		valueBytes := []byte(tmpData.Value)
		if sendArray != nil && len(sendArray)+len(valueBytes) > 200 {
			packDataHexStr := hex.EncodeToString(sendArray)
			cmd, timeoutMs, err := fn(composer, m.CMD_MODIFY_VAR_VALUE, m.CmdData{Type: m.DATA_TYPE_STR, Data: fmt.Sprintf("%s", packDataHexStr)})
			if err != nil {
				return &ScaleRespMsg{m.MODIFY_VAR_RESP, "fail", s.Id}, err
			}

			if res, err := perfCmdNwaitResult(s, cmd, m.MODIFY_VAR_RESP, timeoutMs); err != nil {
				return &ScaleRespMsg{m.MODIFY_VAR_RESP, "fail", s.Id}, err
			} else if res.MsgBody != "ok" {
				return &ScaleRespMsg{m.MODIFY_VAR_RESP, "fail", s.Id}, fmt.Errorf("write eeprom fail")
			}

			i--
			sendArray = nil

		} else {
			binary.BigEndian.PutUint16(idByte, uint16(tmpData.ID))
			sendArray = append(sendArray, idByte...)
			binary.BigEndian.PutUint16(dataLen, uint16(len(tmpData.Value)))
			sendArray = append(sendArray, dataLen...)
			sendArray = append(sendArray, valueBytes...)
			if i == len(headerList.ListData)-1 {
				packDataHexStr := hex.EncodeToString(sendArray)
				cmd, timeoutMs, err := fn(composer, m.CMD_MODIFY_VAR_VALUE, m.CmdData{Type: m.DATA_TYPE_STR, Data: fmt.Sprintf("%s", packDataHexStr)})
				if err != nil {
					return &ScaleRespMsg{m.MODIFY_VAR_RESP, "fail", s.Id}, err
				}
				for _, b := range cmd {
					fmt.Printf("%02x ", b) // 打印每个字节的 16 进制表示并用空格分隔
				}
				// 发送数据包

				if res, err := perfCmdNwaitResult(s, cmd, m.MODIFY_VAR_RESP, timeoutMs); err != nil {
					return &ScaleRespMsg{m.MODIFY_VAR_RESP, "fail", s.Id}, err
				} else if res.MsgBody != "ok" {
					return &ScaleRespMsg{m.MODIFY_VAR_RESP, "fail", s.Id}, fmt.Errorf("write eeprom fail")
				}

			}

		}

	}

	return &ScaleRespMsg{m.MODIFY_VAR_RESP, "ok", s.Id}, nil
}

func StringToFloat64Bytes(s string) []byte {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	bits := math.Float64bits(f)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)
	return bytes
}

func ipv4StringToBytes(ipStr string) []byte {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil
	}

	ip = ip.To4()
	if ip == nil {
		return nil
	}

	return []byte(ip)
}
func intToLittleEndianBytes(dataStr string, size int) ([]byte, error) {
	num, err := strconv.Atoi(dataStr)
	if err != nil {
		return []byte{}, nil
	}
	result := make([]byte, size)
	if size == 2 {
		binary.LittleEndian.PutUint16(result, uint16(num))
		return result, nil
	} else if size == 4 {
		binary.LittleEndian.PutUint32(result, uint32(num))
		return result, nil
	} else if size == 8 {
		binary.LittleEndian.PutUint64(result, uint64(num))
		return result, nil
	}
	return result, nil

}

func isASCII(s string) bool {
	for _, char := range s {
		if char > 127 {
			return false
		}
	}
	return true
}
func stringToLittleEndianBytes(dataStr string, size int) []byte {
	runes := []rune(dataStr)
	result := make([]byte, 0, size)

	if isASCII(dataStr) {
		for i, r := range runes {
			if i >= size {
				break
			}
			result = append(result, byte(r))
		}

	} else {
		for i := 0; i < len(runes); i++ {
			if i*2+1 >= size {
				break
			}
			b := make([]byte, 2)
			binary.LittleEndian.PutUint16(b, uint16(runes[i]))
			result = append(result, b...)
		}
	}

	return result
}

func calculateMD5(input string) [16]byte {
	hash := md5.Sum([]byte(input)) // 计算 MD5 校验值
	return hash                    // 转换为十六进制并返回
}

func openFactory(c *Scale) (*ScaleRespMsg, error, bool) {
	res := false
	composer := c.composer
	fn := composer.ComposeCmd
	reqMsg, _ := excuteSimpCmd(c, m.CMD_CHECK_FAC_MODE, m.EN_FAC_MODE_RESP)
	if reqMsg.MsgBody == "ok" {
		return &ScaleRespMsg{}, nil, true
	}

	var dataStruct FIFromScale

	cmd, timeoutMs, err := fn(composer, m.CMD_GET_FACTORY_INFO, m.CmdData{})
	if err != nil {
		return &ScaleRespMsg{}, err, res
	}
	if res, err := perfCmdNwaitResult(c, cmd, m.GET_FACTORY_INFO_RESP, timeoutMs); err != nil {
		return &ScaleRespMsg{}, err, false
	} else if res.MsgBody == "" {
		return &ScaleRespMsg{}, fmt.Errorf("enable factory mode fail"), false
	} else {
		msgBodyStr, ok := res.MsgBody.(string)
		if !ok {
			return &ScaleRespMsg{}, fmt.Errorf("enable factory mode fail"), false
		}
		err := json.UnmarshalFromString(msgBodyStr, &dataStruct)
		if err != nil {
			return &ScaleRespMsg{}, fmt.Errorf("enable factory mode fail"), false
		}
	}
	if dataStruct.ModelName == "" || dataStruct.ScaleSn == "" {
		return &ScaleRespMsg{}, fmt.Errorf("enable factory mode fail"), false
	}

	crc16Byte := calculateMD5(dataStruct.ModelName + dataStruct.ScaleSn + MD5SEED)
	byte4Md5 := crc16Byte[0:5]
	fmt.Println(byte4Md5)
	//------拿到随机数
	var nums []uint8
	cmd, timeoutMs, err = fn(composer, m.CMD_GET_RANDOM_DATA, m.CmdData{})
	if err != nil {
		return &ScaleRespMsg{}, err, false
	}
	if res, err := perfCmdNwaitResult(c, cmd, m.GET_RANDOM_DATA_RESP, timeoutMs); err != nil {
		return &ScaleRespMsg{}, err, false
	} else if res.MsgBody == "" {
		return &ScaleRespMsg{}, fmt.Errorf("enable factory mode fail"), false
	} else {
		numsInt, ok := res.MsgBody.([]uint8)
		nums = numsInt
		if !ok || len(numsInt) != 2 {
			return &ScaleRespMsg{}, fmt.Errorf("enable factory mode fail"), false
		}
	}

	dataLast := make([]byte, 6)
	dataLast[0] = nums[0]
	dataLast[1] = nums[1]
	copy(dataLast[2:6], byte4Md5)

	//打开工厂模式
	packDataHexStr := hex.EncodeToString(dataLast)
	cmd, timeoutMs, err = fn(composer, m.CMD_EN_FACTORY_MODE, m.CmdData{Type: m.DATA_TYPE_STR, Data: packDataHexStr})
	if err != nil {
		return &ScaleRespMsg{}, err, false
	}
	if res, err := perfCmdNwaitResult(c, cmd, m.EN_FACTORY_MODE_RESP, timeoutMs); err != nil {
		return &ScaleRespMsg{}, err, false
	} else if res.MsgBody != "ok" {
		return &ScaleRespMsg{}, fmt.Errorf("enable factory mode fail"), false
	}

	res = true
	return &ScaleRespMsg{}, err, res

}

// func ReqDownPrnFmt(c *Scale, csvPrnFmt string, seqno string) error {
func ReqDownPrnFmt(c *Scale, req SRequest) (*ScaleRespMsg, error) {
	reg, err, res := openFactory(c)
	if err != nil || !res {
		return reg, err
	}
	composer := c.composer
	fn := composer.ComposeCmd

	reqMsg, _ := excuteSimpCmd(c, m.CMD_GET_SCALE_INFO, m.GET_SCALE_INFO_RESP)
	var scaleInfo SIFromScale
	prnFmtAddr := 0
	prnFmtMaxLenth := 0
	eraseLen := 0
	strData := reqMsg.MsgBody
	if str, ok := strData.(string); ok {
		if err := json.UnmarshalFromString(str, &scaleInfo); err != nil {
			l.Log.Error(err)
			return &ScaleRespMsg{}, fmt.Errorf("get print format address fail")
		}
	} else {
		return &ScaleRespMsg{}, fmt.Errorf("get print format address fail")
	}
	siAddrInfos := scaleInfo.AddrInfos
	fmt.Println(siAddrInfos)
	for _, jsonStr := range siAddrInfos {
		addrInfo, err := parseJSON(jsonStr)
		if err != nil {
			return &ScaleRespMsg{}, fmt.Errorf("get print format address fail")
		}
		if addrInfo.Type == SI_FREE_PRN_INFO {
			prnFmtAddr = addrInfo.Addr
			prnFmtMaxLenth = addrInfo.Lenth
			eraseLen = addrInfo.EraseLen
		}
	}
	if prnFmtAddr == 0 || prnFmtMaxLenth == 0 || eraseLen == 0 {
		return &ScaleRespMsg{}, fmt.Errorf("get plu address fail")
	}
	var reqData ReqPrnData
	if err := json.UnmarshalFromString(req.ReqData, &reqData); err != nil {
		return &ScaleRespMsg{}, err
	}
	for _, file := range reqData.FilePaths {
		fileOrderNo := file[0:1]
		file = file[1:]

		csvFmtContent, err := os.ReadFile(file)
		if err != nil {
			return &ScaleRespMsg{}, err
		}

		tempStr := decryptCsv(string(csvFmtContent))
		if !strings.Contains(tempStr, "ROTATE") {
			return &ScaleRespMsg{}, fmt.Errorf("format error,download fail!")
		}

		if prnfmt.ParserFmtToFile(tempStr, reqData.PrinterModel, eraseLen) {
			// 读取bin文件
			data, err := os.ReadFile("formatBin.bin")
			if err != nil {
				l.Log.Fatal(err)
			}
			// 擦除原本秤上的打印格式
			l.Log.Debug("erase flash on scale")
			no, err := strconv.Atoi(fileOrderNo)
			if err != nil {
				return &ScaleRespMsg{}, err
			}
			addr := eraseLen*(no-1) + prnFmtAddr
			size := eraseLen
			if addr > prnFmtAddr+prnFmtMaxLenth {
				return &ScaleRespMsg{}, err
			}

			loopCnt := size / eraseLen
			addrInLoop := addr
			for i := 0; i < loopCnt; i++ {
				cmd, timeoutMs, err := composer.ComposeCmd(composer, m.CMD_ERASE_FLASH, m.CmdData{Type: m.DATA_TYPE_INT, Data: addrInLoop})
				if err != nil {
					return &ScaleRespMsg{}, err
				}
				if res, err := perfCmdNwaitResult(c, cmd, m.ERASE_FLASH_RESP, timeoutMs); err != nil {
					return &ScaleRespMsg{}, err
				} else if res.MsgBody != "ok" {
					return &ScaleRespMsg{}, fmt.Errorf("erase fail")
				}
				addrInLoop += eraseLen
			}
			// 计算数据包数量
			packetCount := len(data) / DATA_LENGTH_256_TMAX
			if len(data)%DATA_LENGTH_256_TMAX != 0 {
				packetCount += 1
			}
			// 遍历所有数据包
			l.Log.Debug("send data package to scale")
			for i := 0; i < packetCount; i++ {
				// 计算本包数据
				start := i * DATA_LENGTH_256_TMAX
				end := start + DATA_LENGTH_256_TMAX
				if end > len(data) {
					end = len(data)
				}
				packetData := data[start:end]

				// 构建数据包
				// dataPackCmd := buildSendDataPacket(addr, packetData)
				packDataHexStr := hex.EncodeToString(packetData)
				cmd, timeoutMs, err := fn(composer, m.CMD_WRITE_FLASH_256, m.CmdData{Type: m.DATA_TYPE_STR, Data: fmt.Sprintf("%08x:%s", addr, packDataHexStr)})
				if err != nil {
					return &ScaleRespMsg{}, err
				}
				// 发送数据包
				if res, err := perfCmdNwaitResult(c, cmd, m.WRITE_DATA_FLASH_RESP, timeoutMs); err != nil {
					return &ScaleRespMsg{}, err
				} else if res.MsgBody != "ok" {
					return &ScaleRespMsg{}, fmt.Errorf("enable factory mode fail")
				}
				// 地址自增
				addr += 0x100
			}
			l.Log.Info("send bin ok")
		}
	}
	return &ScaleRespMsg{m.DOWN_PRN_FMT_RESP, "ok", c.Id}, nil
}

func getAddrFromScale(c *Scale, typeInt int) (SIAddrInfos, bool) {
	reqMsg, _ := excuteSimpCmd(c, m.CMD_GET_SCALE_INFO, m.GET_SCALE_INFO_RESP)
	var scaleInfo SIFromScale
	var addrInfo SIAddrInfos

	strData := reqMsg.MsgBody
	if str, ok := strData.(string); ok {
		if err := json.UnmarshalFromString(str, &scaleInfo); err != nil {
			l.Log.Error(err)
			return addrInfo, false
		}
	} else {
		return addrInfo, false
	}
	siAddrInfos := scaleInfo.AddrInfos
	fmt.Println(siAddrInfos)
	for _, jsonStr := range siAddrInfos {
		addrInfo, err := parseJSON(jsonStr)
		if err != nil {
			return addrInfo, false
		}
		if addrInfo.Type == typeInt {
			return addrInfo, true
		}
	}
	return addrInfo, false

}

func ReqDownDefaultPrnFmt(c *Scale, req SRequest) (*ScaleRespMsg, error) {
	reg, err, res := openFactory(c)
	if err != nil || !res {
		return reg, err
	}

	prnFmtAddr := 0
	prnFmtMaxLenth := 0
	eraseLen := 0
	addrInfo, res := getAddrFromScale(c, SI_DEF_PRN_INFO)
	if !res {
		return &ScaleRespMsg{}, fmt.Errorf("get plu address fail")
	}

	prnFmtAddr = addrInfo.Addr
	prnFmtMaxLenth = addrInfo.Lenth
	eraseLen = addrInfo.EraseLen

	if prnFmtAddr == 0 || prnFmtMaxLenth == 0 || eraseLen == 0 {
		return &ScaleRespMsg{}, fmt.Errorf("get plu address fail")
	}
	var reqData ReqDefaultPrnData
	if err := json.UnmarshalFromString(req.ReqData, &reqData); err != nil {
		return &ScaleRespMsg{}, err
	}

	zipFilePath := reqData.FilePath
	//将所有文件解密后的内容和md5都存储到数组里
	strFileDataArray, strMd5Array, res := getFileDataAndMd5(zipFilePath)
	if !res {
		return &ScaleRespMsg{}, fmt.Errorf("fail,file error")
	}

	//检查所有文件的md5是否正确
	if !checkFilesMd5(strFileDataArray, strMd5Array) {
		return &ScaleRespMsg{}, fmt.Errorf("fail,file error")
	}

	//分析打印格式
	composer := c.composer
	fn := composer.ComposeCmd
	if prnfmt.ParserDefFmtToFile(strFileDataArray, reqData.PrinterModel, prnFmtMaxLenth) {
		// 读取bin文件
		data, err := os.ReadFile("formatBin.bin")
		if err != nil {
			l.Log.Fatal(err)
		}
		// 擦除原本秤上的打印格式
		l.Log.Debug("erase flash on scale")

		addr := prnFmtAddr
		size := prnFmtMaxLenth

		loopCnt := size / eraseLen
		addrInLoop := addr
		for i := 0; i < loopCnt; i++ {
			cmd, timeoutMs, err := composer.ComposeCmd(composer, m.CMD_ERASE_FLASH, m.CmdData{Type: m.DATA_TYPE_INT, Data: addrInLoop})
			if err != nil {
				return &ScaleRespMsg{}, err
			}
			if res, err := perfCmdNwaitResult(c, cmd, m.ERASE_FLASH_RESP, timeoutMs); err != nil {
				return &ScaleRespMsg{}, err
			} else if res.MsgBody != "ok" {
				return &ScaleRespMsg{}, fmt.Errorf("erase fail")
			}
			addrInLoop += eraseLen
		}
		// 计算数据包数量
		packetCount := len(data) / DATA_LENGTH_256_TMAX
		if len(data)%DATA_LENGTH_256_TMAX != 0 {
			packetCount += 1
		}
		// 遍历所有数据包
		l.Log.Debug("send data package to scale")
		for i := 0; i < packetCount; i++ {
			// 计算本包数据
			start := i * DATA_LENGTH_256_TMAX
			end := start + DATA_LENGTH_256_TMAX
			if end > len(data) {
				end = len(data)
			}
			packetData := data[start:end]

			// 构建数据包
			// dataPackCmd := buildSendDataPacket(addr, packetData)
			packDataHexStr := hex.EncodeToString(packetData)
			cmd, timeoutMs, err := fn(composer, m.CMD_WRITE_FLASH_256, m.CmdData{Type: m.DATA_TYPE_STR, Data: fmt.Sprintf("%08x:%s", addr, packDataHexStr)})
			if err != nil {
				return &ScaleRespMsg{}, err
			}
			// 发送数据包
			if res, err := perfCmdNwaitResult(c, cmd, m.WRITE_DATA_FLASH_RESP, timeoutMs); err != nil {
				return &ScaleRespMsg{}, err
			} else if res.MsgBody != "ok" {
				return &ScaleRespMsg{}, fmt.Errorf("enable factory mode fail")
			}
			// 地址自增
			addr += 0x100
		}
		l.Log.Info("send bin ok")
	} else {
		return &ScaleRespMsg{}, fmt.Errorf("fail,Check for over 150 variables, excessive 8K print format, or incorrect print format.")
	}

	return &ScaleRespMsg{m.DOWN_DEFAULT_PRN_FMT_RESP, "ok", c.Id}, nil
}

func getFileDataAndMd5(zipFilePath string) ([12]string, [12]string, bool) {
	var strFileDataArray [12]string
	var strMd5Array [12]string
	r, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return strFileDataArray, strMd5Array, false
	}
	defer r.Close()
	for _, f := range r.File {
		fmt.Println("File:", f.Name)
		// 打开文件
		rc, err := f.Open()
		if err != nil {
			return strFileDataArray, strMd5Array, false
		}
		defer rc.Close()
		num, res := getFileNum(f.Name)
		if !res {
			return strFileDataArray, strMd5Array, false
		}
		tempContent, err := io.ReadAll(rc)
		if err != nil {
			return strFileDataArray, strMd5Array, false
		}
		if strings.Contains(f.Name, "fmt") {
			tempStr := decryptCsv(string(tempContent))
			if !strings.Contains(tempStr, "ROTATE") {
				return strFileDataArray, strMd5Array, false
			}
			strFileDataArray[num-1] = tempStr
		} else if strings.Contains(f.Name, "txt") {
			strMd5Array[num-1] = string(tempContent)
		}
	}
	return strFileDataArray, strMd5Array, true
}

func checkFilesMd5(strFileDataArray [12]string, strMd5Array [12]string) bool {
	for i := 0; i < 12; i++ {
		if strFileDataArray[i] != "" {
			tempMd5Arr := calculateMD5(strFileDataArray[i] + MD5SEED) //统一为seed,要改打包的UI，此处先暂定这样
			var md5TmpStr string
			for _, b := range tempMd5Arr {
				md5TmpStr += fmt.Sprintf("%02X", b) // 将每个字节转换为两位16进制格式的字符串并拼接
			}
			if !strings.EqualFold(md5TmpStr, strMd5Array[i]) {
				return false
			}
		}
	}
	return true
}

func getFileNum(fNameStr string) (int, bool) {
	re := regexp.MustCompile(`(\d+)`) // 使用正则表达式提取数字部分
	match := re.FindStringSubmatch(fNameStr)

	if len(match) > 1 {
		numStr := match[1]               // 提取到的数字部分
		num, err := strconv.Atoi(numStr) // 将字符串转换为int类型
		if err == nil {
			return num, true
		} else {
			return 0, false
		}
	}
	return 0, false
}

func decryptCsv(encryptedCsv string) string {
	var decryptedBytes []byte
	encryptedBytes := []byte(encryptedCsv)

	for i := 0; i < len(encryptedBytes); i += 2 {
		decryptedByte := (int32(encryptedBytes[i])-3)<<4 | (int32(encryptedBytes[i+1])-3)&0x0F
		decryptedBytes = append(decryptedBytes, byte(decryptedByte))
	}

	return string(decryptedBytes)
}

func ReqInsertPlu(c *Scale, req SRequest) (*ScaleRespMsg, error) {
	var reqData ReqPluData

	if err := json.UnmarshalFromString(req.ReqData, &reqData); err != nil {
		return &ScaleRespMsg{}, err
	}
	var file = reqData.FilePath
	var nameMaxLen = reqData.NameMaxLen
	var headBytes []byte
	composer := c.composer

	l.Log.Debug("get plu address to scale")
	cmd, timeoutMs, err := composer.ComposeCmd(composer, m.CMD_GET_PLU_HEAD, m.CmdData{})
	if err != nil {
		return &ScaleRespMsg{}, err
	}

	if res, err := perfCmdNwaitResult(c, cmd, m.READ_FLASH_DATA_RESP, timeoutMs); err != nil {
		return &ScaleRespMsg{}, err
	} else if res.MsgBody == "fail" {
		return &ScaleRespMsg{}, fmt.Errorf("enable factory mode fail")
	} else {
		strData := res.MsgBody
		if str, ok := strData.(string); ok {
			addrBytes := []byte(str)
			if len(addrBytes) >= 100 {
				for i := 0; i < len(addrBytes); i++ {
					if i+2 < len(addrBytes) {
						if addrBytes[i] == 0xff && addrBytes[i+1] == 0x01 && addrBytes[i+2] == 0x00 {
							headBytes = addrBytes[35 : i+3]
						}
					}
				}
				if len(headBytes) == 0 {
					return &ScaleRespMsg{}, fmt.Errorf("get plu head fail")
				}

				//前4个是起始地址，中间4个字节是现在可用的地址，最后四个是可用空间大小
				// insertPluAddr = int(binary.BigEndian.Uint32(addrBytes[4:8]))
				// insertPluSize = int(binary.BigEndian.Uint32(addrBytes[8:12]))
			} else {
				return &ScaleRespMsg{}, fmt.Errorf("get plu head fail")
			}
		} else {
			return &ScaleRespMsg{}, fmt.Errorf("get plu head fail")
		}
	}

	bufPluNumData, insertHeadBytes, resParser := ParserInsertPlu(string(file), nameMaxLen)

	if !bytes.Equal(headBytes, insertHeadBytes) {
		return &ScaleRespMsg{}, fmt.Errorf("plu head is inconsistent,fail")
	}

	if resParser {
		// 读取bin文件
		data, err := os.ReadFile("plu.bin")
		if err != nil {
			l.Log.Fatal(err)
		}
		l.Log.Debug("send enable factory mode cmd to scale")

		reg, err, res := openFactory(c)
		if err != nil || !res {
			return reg, err
		}
		composer := c.composer
		fn := composer.ComposeCmd

		l.Log.Debug("get plu address to scale")
		cmd, timeoutMs, err = composer.ComposeCmd(composer, m.CMD_INSERT_PLU_ADDR, m.CmdData{})
		if err != nil {
			return &ScaleRespMsg{}, err
		}

		var insertPluAddr int
		var insertPluSize int
		if res, err := perfCmdNwaitResult(c, cmd, m.INSERT_PLU_ADDR_RESP, timeoutMs); err != nil {
			return &ScaleRespMsg{}, err
		} else if res.MsgBody == "fail" {
			return &ScaleRespMsg{}, fmt.Errorf("enable factory mode fail")
		} else {
			strData := res.MsgBody
			if str, ok := strData.(string); ok {
				addrBytes := []byte(str)
				if len(addrBytes) >= 12 {
					//前4个是起始地址，中间4个字节是现在可用的地址，最后四个是可用空间大小
					insertPluAddr = int(binary.BigEndian.Uint32(addrBytes[4:8]))
					insertPluSize = int(binary.BigEndian.Uint32(addrBytes[8:12]))
				} else {
					insertPluAddr = 0
					insertPluSize = 0
					return &ScaleRespMsg{}, fmt.Errorf("get plu address fail")
				}

			} else {
				return &ScaleRespMsg{}, fmt.Errorf("get plu address fail")
			}
		}

		// 比较空间大小
		addr := insertPluAddr
		totalSize := insertPluSize
		size := len(data)
		if totalSize < size {
			return &ScaleRespMsg{}, fmt.Errorf("no enough space to insert")
		}
		if len(bufPluNumData) > 256 {
			return &ScaleRespMsg{}, fmt.Errorf("The number cannot be more than 60.")
		}
		l.Log.Debug("delete plu number cmd to scale")
		cmd, timeoutMs, err = c.composer.ComposeCmd(c.composer, m.CMD_DEL_PLU, m.CmdData{Type: m.DATA_TYPE_STR, Data: string(bufPluNumData)})
		if err != nil {
			return &ScaleRespMsg{}, err
		}

		if res, err := perfCmdNwaitResult(c, cmd, m.DEL_PLU_RESP, timeoutMs); err != nil {
			return &ScaleRespMsg{}, err
		} else if res.MsgBody != "ok" {
			return &ScaleRespMsg{}, fmt.Errorf("delete plu id fail")
		}

		packetCount := len(data) / DATA_LENGTH_256_TMAX
		if len(data)%DATA_LENGTH_256_TMAX != 0 {
			packetCount += 1
		}
		l.Log.Debug("send data package to scale")
		for i := 0; i < packetCount; i++ {
			// 计算本包数据
			start := i * DATA_LENGTH_256_TMAX
			end := start + DATA_LENGTH_256_TMAX
			if end > len(data) {
				end = len(data)
			}
			packetData := data[start:end]
			packDataHexStr := hex.EncodeToString(packetData)
			cmd, timeoutMs, err = fn(composer, m.CMD_WRITE_FLASH_256, m.CmdData{Type: m.DATA_TYPE_STR, Data: fmt.Sprintf("%08x:%s", addr, packDataHexStr)})
			if err != nil {
				return &ScaleRespMsg{}, err
			}
			if res, err := perfCmdNwaitResult(c, cmd, m.WRITE_DATA_FLASH_RESP, timeoutMs); err != nil {
				return &ScaleRespMsg{}, err
			} else if res.MsgBody != "ok" {
				return &ScaleRespMsg{}, fmt.Errorf("enable factory mode fail")
			}
			addr += DATA_LENGTH_256_TMAX
		}
		l.Log.Info("send bin ok")
	}
	return &ScaleRespMsg{m.INSERT_PLU_RESP, "ok", c.Id}, nil
}

func ReqDownPlu(c *Scale, req SRequest) (*ScaleRespMsg, error) {
	var reqData ReqPluData
	if err := json.UnmarshalFromString(req.ReqData, &reqData); err != nil {
		return &ScaleRespMsg{}, err
	}
	var file = reqData.FilePath
	var nameMaxLen = reqData.NameMaxLen

	if ParserPluFile(string(file), nameMaxLen) {
		// 读取bin文件
		data, err := os.ReadFile("plu.bin")
		if err != nil {
			l.Log.Fatal(err)
		}
		l.Log.Debug("send enable factory mode cmd to scale")
		reg, err, res := openFactory(c)
		if err != nil || !res {
			return reg, err
		}
		composer := c.composer
		fn := composer.ComposeCmd

		//获取秤上的PLU地址
		reqMsg, _ := excuteSimpCmd(c, m.CMD_GET_SCALE_INFO, m.GET_SCALE_INFO_RESP)
		var scaleInfo SIFromScale
		pluStartAddr := 0
		pluMaxLenth := 0

		strData := reqMsg.MsgBody
		if str, ok := strData.(string); ok {
			if err := json.UnmarshalFromString(str, &scaleInfo); err != nil {
				l.Log.Error(err)
				return &ScaleRespMsg{}, fmt.Errorf("get plu address fail")
			}
		} else {
			return &ScaleRespMsg{}, fmt.Errorf("get plu address fail")
		}

		siAddrInfos := scaleInfo.AddrInfos
		fmt.Println(siAddrInfos)

		for _, jsonStr := range siAddrInfos {
			addrInfo, err := parseJSON(jsonStr)
			if err != nil {
				return &ScaleRespMsg{}, fmt.Errorf("get plu address fail")
			}

			if addrInfo.Type == SI_PLU_INFO {
				pluStartAddr = addrInfo.Addr
				pluMaxLenth = addrInfo.Lenth
			}
		}
		if pluStartAddr == 0 || pluMaxLenth == 0 {
			return &ScaleRespMsg{}, fmt.Errorf("get plu address fail")
		}
		// 擦除原本秤上的PLU
		// addr, size := mycmd.GetPluRomAddrNSize(c.ScaleCat)
		addr := pluStartAddr
		maxSize := pluMaxLenth
		size := len(data)
		if maxSize < size {
			return &ScaleRespMsg{}, fmt.Errorf("no enough space to download")
		}
		loopCnt := size/4096 + 1
		addrInLoop := addr
		for i := 0; i < loopCnt; i++ {
			cmd, timeoutMs, err := composer.ComposeCmd(composer, m.CMD_ERASE_FLASH, m.CmdData{Type: m.DATA_TYPE_INT, Data: addrInLoop})
			if err != nil {
				return &ScaleRespMsg{}, err
			}
			if res, err := perfCmdNwaitResult(c, cmd, m.ERASE_FLASH_RESP, timeoutMs); err != nil {
				return &ScaleRespMsg{}, err
			} else if res.MsgBody != "ok" {
				return &ScaleRespMsg{}, fmt.Errorf("erase fail")
			}
			addrInLoop += 4096
		}
		packetCount := len(data) / DATA_LENGTH_256_TMAX
		if len(data)%DATA_LENGTH_256_TMAX != 0 {
			packetCount += 1
		}
		l.Log.Debug("send data package to scale")
		for i := 0; i < packetCount; i++ {
			// 计算本包数据
			start := i * DATA_LENGTH_256_TMAX
			end := start + DATA_LENGTH_256_TMAX
			if end > len(data) {
				end = len(data)
			}
			packetData := data[start:end]
			packDataHexStr := hex.EncodeToString(packetData)
			cmd, timeoutMs, err := fn(composer, m.CMD_WRITE_FLASH_256, m.CmdData{Type: m.DATA_TYPE_STR, Data: fmt.Sprintf("%08x:%s", addr, packDataHexStr)})
			if err != nil {
				return &ScaleRespMsg{}, err
			}
			if res, err := perfCmdNwaitResult(c, cmd, m.WRITE_DATA_FLASH_RESP, timeoutMs); err != nil {
				return &ScaleRespMsg{}, err
			} else if res.MsgBody != "ok" {
				return &ScaleRespMsg{}, fmt.Errorf("enable factory mode fail")
			}
			addr += DATA_LENGTH_256_TMAX
		}
		l.Log.Info("send bin ok")
	} else {
		return &ScaleRespMsg{}, fmt.Errorf("fail,data error")
	}
	//备份PLU表格 记录md5值和文件的对应关系
	destPath := getPluFilePath(getCurrPath())
	CopyFile(string(file), destPath) //备份plu下发的数据
	md5Str := fileToMd5(destPath)
	var pluDown PluRec
	pluDown.FileName = destPath
	pluDown.Md5 = md5Str
	c.AddPluDownRec(pluDown)
	l.Log.Info("erase insert addr")
	//下命令让秤擦除新增区
	composer := c.composer
	cmd, timeoutMs, err := composer.ComposeCmd(composer, m.CMD_ERASE_INSERT_PLU, m.CmdData{})
	if err != nil {
		return &ScaleRespMsg{}, err
	}

	if res, err := perfCmdNwaitResult(c, cmd, m.ERASE_INSERT_PLU_RESP, timeoutMs); err != nil {
		return &ScaleRespMsg{}, err
	} else if res.MsgBody == "fail" {
		return &ScaleRespMsg{}, fmt.Errorf("erase insert plu fail")
	}
	return &ScaleRespMsg{m.DOWN_PLU_RESP, "ok", c.Id}, nil
}

func parseJSON(jsonStr string) (SIAddrInfos, error) {
	var addrInfo SIAddrInfos
	err := json.Unmarshal([]byte(jsonStr), &addrInfo)
	return addrInfo, err
}

func getPluFilePath(currPath string) string {
	timestamp := time.Now().Unix()
	timeStr := strconv.FormatInt(timestamp, 10)
	currPath = currPath + string(os.PathSeparator) + timeStr + ".xlsx"
	return currPath
}

func getCurrPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))
	currentPath := path[:index]
	currentPath = filepath.Join(currentPath, m.SRV_DATA_PATH)
	pluBackPath := filepath.Join(currentPath, m.PLU_BACK_PATH)
	if _, err := os.Stat(pluBackPath); os.IsNotExist(err) {
		os.MkdirAll(pluBackPath, os.ModePerm)
	}
	currentPath = filepath.Join(currentPath, m.PLU_BACK_PATH)
	return currentPath
}

func CopyFile(srcFilePath, dstFilePath string) (written int64, err error) {

	srcfile, err := os.Open(srcFilePath)
	defer func() {
		if err = srcfile.Close(); err != nil {
			l.Log.Info(err)
		}
	}()
	if err != nil {
		l.Log.Info(err)
	}
	// 构建 src Reader
	srcReaer := bufio.NewReader(srcfile)

	//打开dstFilePath
	dstfile, err := os.OpenFile(dstFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	defer func() {
		if err := dstfile.Close(); err != nil {
			l.Log.Info(err)
		}
	}()
	if err != nil {
		l.Log.Info(err)
	}
	// 构建 src Writer
	dstWriter := bufio.NewWriter(dstfile)

	return io.Copy(dstWriter, srcReaer)

}

func ReqSetScaleTime(c *Scale, req SRequest) (*ScaleRespMsg, error) {
	reqData := req.ReqData
	num, err := strconv.ParseUint(reqData, 10, 32)
	if err != nil {
		return &ScaleRespMsg{m.SET_SCALE_TIME_RESP, "fail, data error", c.Id}, nil
	}
	reg, err, res := openFactory(c)
	if err != nil || !res {
		return reg, err
	}

	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, uint32(num))
	l.Log.Debug("send cmd to scale")
	cmd, timeoutMs, err := c.composer.ComposeCmd(c.composer, m.CMD_SET_SCALE_TIME, m.CmdData{Type: m.DATA_TYPE_STR, Data: string(bytes)})
	if err != nil {
		return &ScaleRespMsg{}, err
	}
	if res, err := perfCmdNwaitResult(c, cmd, m.SET_SCALE_TIME_RESP, timeoutMs); err != nil {
		return &ScaleRespMsg{}, err
	} else if res.MsgBody != "ok" {
		return &ScaleRespMsg{}, fmt.Errorf("set time fail")
	}
	return &ScaleRespMsg{m.SET_SCALE_TIME_RESP, "ok", c.Id}, nil
}

func ReqDelPlu(c *Scale, req SRequest) (*ScaleRespMsg, error) {
	var reqData ReqDelPLuData
	if err := json.UnmarshalFromString(req.ReqData, &reqData); err != nil {
		return &ScaleRespMsg{}, err
	}

	if len(reqData.PluId) == 0 {
		return &ScaleRespMsg{m.DEL_PLU_RESP, "fail, no PLU id", c.Id}, nil
	}

	bufData, _ := ParserDelPlu(reqData.PluId)

	l.Log.Debug("send cmd to scale")
	cmd, timeoutMs, err := c.composer.ComposeCmd(c.composer, m.CMD_DEL_PLU, m.CmdData{Type: m.DATA_TYPE_STR, Data: string(bufData)})
	if err != nil {
		return &ScaleRespMsg{}, err
	}

	if res, err := perfCmdNwaitResult(c, cmd, m.DEL_PLU_RESP, timeoutMs); err != nil {
		return &ScaleRespMsg{}, err
	} else if res.MsgBody != "ok" {
		return &ScaleRespMsg{}, fmt.Errorf("delete plu id fail")
	}

	return &ScaleRespMsg{m.DEL_PLU_RESP, "ok", c.Id}, nil
}

func deleteFile(fileName string) error {
	currentDir, err := os.Getwd()
	if err != nil {
		l.Log.Debug("failed get current path")
		return nil
	}
	// 拼接文件的完整路径
	fullPath := filepath.Join(currentDir, fileName)
	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		l.Log.Debug("file does not exist ")
		return nil
	}
	// 删除文件
	err = os.Remove(fullPath)
	if err != nil {
		l.Log.Debug("delete file failed")
		return err
	}
	return nil
}

func ReqGetWeightErr(c *Scale) (*ScaleRespMsg, error) {
	l.Log.Debug("send enable factory mode cmd to scale")
	reg, err, res := openFactory(c)
	if err != nil || !res {
		return reg, err
	}

	//获取秤上的OLUL地址
	reqMsg, err := excuteSimpCmd(c, m.CMD_GET_SCALE_INFO, m.GET_SCALE_INFO_RESP)
	var scaleInfo SIFromScale
	olAddr := 0
	ulAddr := 0

	strData := reqMsg.MsgBody
	if str, ok := strData.(string); ok {
		if err := json.UnmarshalFromString(str, &scaleInfo); err != nil {
			l.Log.Error(err)
			return &ScaleRespMsg{}, fmt.Errorf("get ol/ul address fail")
		}

	} else {
		return &ScaleRespMsg{}, fmt.Errorf("get ol/ul address fail")
	}

	siAddrInfos := scaleInfo.AddrInfos
	fmt.Println(siAddrInfos)

	for _, jsonStr := range siAddrInfos {
		addrInfo, err := parseJSON(jsonStr)
		if err != nil {
			return &ScaleRespMsg{}, fmt.Errorf("get ol/ul address fail")
		}

		if addrInfo.Type == SI_OL_INFO {
			olAddr = addrInfo.Addr
		}
		if addrInfo.Type == SI_UL_INFO {
			ulAddr = addrInfo.Addr
		}
	}
	if olAddr == 0 || ulAddr == 0 {
		return &ScaleRespMsg{}, fmt.Errorf("get plu address fail")
	}

	byteOLArray, resOL := getOlUlFromScale(c, olAddr)
	if !resOL {
		return &ScaleRespMsg{}, fmt.Errorf("get data fail")
	}
	byteULArray, resUL := getOlUlFromScale(c, ulAddr)
	if !resUL {
		return &ScaleRespMsg{}, fmt.Errorf("get data fail")
	}

	var respInfo OlUlInfo
	if len(byteOLArray) == 8 && len(byteULArray) == 8 {
		respInfo.OlTime = int(binary.LittleEndian.Uint32(byteOLArray[0:4]))
		respInfo.OlCnt = int(binary.LittleEndian.Uint32(byteOLArray[4:8]))
		respInfo.UlTime = int(binary.LittleEndian.Uint32(byteULArray[0:4]))
		respInfo.UlCnt = int(binary.LittleEndian.Uint32(byteULArray[4:8]))
	}

	respInfoStr, err := json.MarshalToString(respInfo)
	if err != nil {
		return &ScaleRespMsg{}, fmt.Errorf("get data fail")
	}

	return &ScaleRespMsg{m.GET_WEIGHT_ERR_RESP, respInfoStr, c.Id}, nil
}

func getOlUlFromScale(c *Scale, addr int) ([]byte, bool) {
	var byteArray []byte
	composer := c.composer
	fn := composer.ComposeCmd
	l.Log.Debug("read flash from scale")
	getData8Len := []byte{0x00, 0x08}
	getData256Len := []byte{0x01, 0x00}
	packDataHexStr := hex.EncodeToString(getData8Len)
	packData256HexStr := hex.EncodeToString(getData256Len)
	totalSpace := (8 * 1024)
	loopCnt := totalSpace / DATA_LENGTH_256_TMAX
	if totalSpace%DATA_LENGTH_256_TMAX != 0 {
		loopCnt += 1
	}
	addrInLoop := addr
	for i := 0; i < loopCnt; i++ {
		cmd, timeoutMs, err := fn(composer, m.CMD_GET_WEIGHT_ERR, m.CmdData{Type: m.DATA_TYPE_STR, Data: fmt.Sprintf("%08x:%s", addrInLoop, packDataHexStr)})
		if err != nil {
			return byteArray, false
		}
		if res, err := perfCmdNwaitResult(c, cmd, m.READ_FLASH_DATA_RESP, timeoutMs); err != nil {
			return byteArray, false
		} else if res.MsgBody == "fail" {
			return byteArray, false
		} else {
			strData := res.MsgBody
			if str, ok := strData.(string); ok {
				addrBytes := []byte(str)
				if len(addrBytes) == 8 &&
					addrBytes[0] == 0xFF && addrBytes[1] == 0xFF && addrBytes[2] == 0xFF && addrBytes[3] == 0xFF &&
					addrBytes[4] == 0xFF && addrBytes[5] == 0xFF && addrBytes[6] == 0xFF && addrBytes[7] == 0xFF {
					// 0到7位置都是 FF
					break
				} else if len(addrBytes) != 8 {
					return byteArray, false
				}
			} else {
				return byteArray, false
			}
		}
		addrInLoop += 256
	}

	if addrInLoop > addr {
		cmd, timeoutMs, err := fn(composer, m.CMD_GET_WEIGHT_ERR, m.CmdData{Type: m.DATA_TYPE_STR, Data: fmt.Sprintf("%08x:%s", addrInLoop-256, packData256HexStr)})
		if err != nil {
			return byteArray, false
		}

		if res, err := perfCmdNwaitResult(c, cmd, m.READ_FLASH_DATA_RESP, timeoutMs); err != nil {
			return byteArray, false
		} else if res.MsgBody == "fail" {
			return byteArray, false
		} else {
			strData := res.MsgBody
			if str, ok := strData.(string); ok {
				addrBytes := []byte(str)
				if len(addrBytes) == 256 {
					len := 0
					for i := 0; i < 256/8; i++ {
						if addrBytes[len] == 0xFF && addrBytes[len+1] == 0xFF && addrBytes[len+2] == 0xFF && addrBytes[len+3] == 0xFF &&
							addrBytes[len+4] == 0xFF && addrBytes[len+5] == 0xFF && addrBytes[len+6] == 0xFF && addrBytes[len+7] == 0xFF {

							byteArray = addrBytes[len-8 : len]
							break

						} else if len+8 == 256 {
							byteArray = addrBytes[256-8 : 256]
							break
						}
						len += 8
					}
				} else if len(addrBytes) != 256 {
					return byteArray, false
				}
			} else {
				return byteArray, false
			}
		}

	} else {
		byteArray = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	}
	return byteArray, true

}

func ReqGetOneEepromInfo(c *Scale, req SRequest) (*ScaleRespMsg, error) {
	var eepromAddr int
	var eepromSize int
	var dataVal int

	if req.ReqData == WIFI_BT_INFO {
		eepromAddr, eepromSize = eeprom.GetFuncAddrSize("M_WIRELESS_PERIPHERAL_TYPE")
	}
	reg, err, res := openFactory(c)
	if err != nil || !res {
		return reg, err
	}
	composer := c.composer

	if eepromSize > 0 && eepromAddr < 256 {

		l.Log.Debug("get eeprom data from scale")
		cmd, timeoutMs, err := composer.ComposeCmd(composer, m.CMD_READ_EEPROM_256, m.CmdData{})
		if err != nil {
			return &ScaleRespMsg{}, err
		}

		if res, err := perfCmdNwaitResult(c, cmd, m.READ_FLASH_DATA_RESP, timeoutMs); err != nil {
			return &ScaleRespMsg{}, err
		} else if res.MsgBody == "fail" {
			return &ScaleRespMsg{}, fmt.Errorf("enable factory mode fail")
		} else {
			strData := res.MsgBody
			if str, ok := strData.(string); ok {
				addrBytes := []byte(str)
				if len(addrBytes) >= 256 {
					if eepromSize == 1 {
						dataVal = int(addrBytes[eepromAddr])
						if dataVal == 0 {
							return &ScaleRespMsg{m.GET_ONE_EEPROM_INFO_RESP, "ok,off", c.Id}, nil
						} else if dataVal == 1 {
							return &ScaleRespMsg{m.GET_ONE_EEPROM_INFO_RESP, "ok,wifi", c.Id}, nil
						} else if dataVal == 2 {
							return &ScaleRespMsg{m.GET_ONE_EEPROM_INFO_RESP, "ok,bt", c.Id}, nil
						}
					}

				}

			} else {
				return &ScaleRespMsg{}, fmt.Errorf("get plu address fail")
			}
		}

	}

	return &ScaleRespMsg{m.GET_ONE_EEPROM_INFO_RESP, "", c.Id}, nil

}

type EData struct {
	FiledName   string `json:"filedName"`
	Size        int    `json:"size"`
	Addr        int    `json:"addr"`
	Type        string `json:"type"`
	SubType     string `json:"subType"`
	Permission  int    `json:"permission"`
	Values      string `json:"values"`
	CurrValue   string `json:"currValue"`
	Description string `json:"description"`
	Comment     string `json:"comment"`
	Category    string `json:"category"`
}

type EepromDataResp struct {
	EepromData []EData
}

func ReqGetAllEepromInfo(c *Scale) (*ScaleRespMsg, error) {

	var filedDataLen = 0
	var eepromResp EepromDataResp
	fieldData, res := eeprom.GetExcelData()
	if !res {
		return &ScaleRespMsg{}, fmt.Errorf("fail,no eeprom data file")
	}
	filedDataLen = len(fieldData.EepromStruct)
	if filedDataLen <= 0 {
		return &ScaleRespMsg{}, fmt.Errorf("fail,eeprom data file is empty")
	}
	for i := 0; i < filedDataLen; i++ {
		var eData EData
		eData.FiledName = fieldData.EepromStruct[i].FieldName
		eData.Permission = fieldData.EepromStruct[i].Permission
		eData.Comment = fieldData.EepromStruct[i].Comments
		eData.Category = fieldData.EepromStruct[i].Category
		eData.Addr = fieldData.EepromStruct[i].Addr
		eData.Size = fieldData.EepromStruct[i].Size
		eData.Description = fieldData.EepromStruct[i].Description
		eData.Values = fieldData.EepromStruct[i].Values
		eData.CurrValue = ""
		eData.Type = fieldData.EepromStruct[i].Type
		eData.SubType = fieldData.EepromStruct[i].SubType
		eepromResp.EepromData = append(eepromResp.EepromData, eData)
	}
	eepromStr, _ := json.MarshalToString(eepromResp.EepromData)
	composer := c.composer

	_, err, res := openFactory(c)
	if err != nil || !res {
		return &ScaleRespMsg{m.GET_ALL_EEPROM_INFO_RESP, eepromStr, c.Id}, nil
	}

	l.Log.Debug("get eeprom data from scale")
	dataBuffer := make([]byte, 0)
	cmd, timeoutMs, err := composer.ComposeCmd(composer, m.CMD_READ_EEPROM_256, m.CmdData{})
	if err != nil {
		return &ScaleRespMsg{m.GET_ALL_EEPROM_INFO_RESP, eepromStr, c.Id}, nil
	}
	if res, err := perfCmdNwaitResult(c, cmd, m.READ_FLASH_DATA_RESP, timeoutMs); err != nil {
		return &ScaleRespMsg{m.GET_ALL_EEPROM_INFO_RESP, eepromStr, c.Id}, nil
	} else if res.MsgBody == "fail" {
		return &ScaleRespMsg{m.GET_ALL_EEPROM_INFO_RESP, eepromStr, c.Id}, nil
	} else {
		strData := res.MsgBody
		if str, ok := strData.(string); ok {
			addrBytes := []byte(str)
			if len(addrBytes) >= 256 {
				dataBuffer = append(dataBuffer, addrBytes...)
			} else {
				return &ScaleRespMsg{m.GET_ALL_EEPROM_INFO_RESP, eepromStr, c.Id}, nil
			}
		} else {
			return &ScaleRespMsg{m.GET_ALL_EEPROM_INFO_RESP, eepromStr, c.Id}, nil
		}
	}
	cmd, timeoutMs, err = composer.ComposeCmd(composer, m.CMD_READ_EEPROM_512, m.CmdData{})
	if err != nil {
		return &ScaleRespMsg{m.GET_ALL_EEPROM_INFO_RESP, eepromStr, c.Id}, nil
	}
	if res, err := perfCmdNwaitResult(c, cmd, m.READ_FLASH_DATA_RESP, timeoutMs); err != nil {
		return &ScaleRespMsg{m.GET_ALL_EEPROM_INFO_RESP, eepromStr, c.Id}, nil
	} else if res.MsgBody == "fail" {
		return &ScaleRespMsg{m.GET_ALL_EEPROM_INFO_RESP, eepromStr, c.Id}, nil
	} else {
		strData := res.MsgBody
		if str, ok := strData.(string); ok {
			addrBytes := []byte(str)
			if len(addrBytes) >= 256 {
				dataBuffer = append(dataBuffer, addrBytes...)
			} else {
				return &ScaleRespMsg{m.GET_ALL_EEPROM_INFO_RESP, eepromStr, c.Id}, nil
			}
		} else {
			return &ScaleRespMsg{m.GET_ALL_EEPROM_INFO_RESP, eepromStr, c.Id}, nil
		}
	}
	eepromResp, res = paserEepromData(dataBuffer, eepromResp, 0)
	eepromStr, _ = json.MarshalToString(eepromResp.EepromData)
	return &ScaleRespMsg{m.GET_ALL_EEPROM_INFO_RESP, eepromStr, c.Id}, nil

}

func paserEepromData(eepromBytes []byte, eepromResp EepromDataResp, startAddr int) (EepromDataResp, bool) {
	var dataAddr int
	var dataSize int
	var dataType string
	var dataSubType string

	for i := 0; i < len(eepromResp.EepromData); i++ {
		dataAddr = eepromResp.EepromData[i].Addr
		dataSize = eepromResp.EepromData[i].Size
		dataType = eepromResp.EepromData[i].Type
		dataSubType = eepromResp.EepromData[i].SubType

		if dataAddr >= startAddr && (dataAddr-startAddr) < len(eepromBytes) && len(eepromBytes) > dataAddr+dataSize {

			if dataType == "int" {
				if dataSize == 1 {
					if eepromBytes[dataAddr] == 0xff {
						eepromResp.EepromData[i].Permission = 0
						eepromResp.EepromData[i].CurrValue = ""
					} else {
						dataInt := int(eepromBytes[dataAddr])
						eepromResp.EepromData[i].CurrValue = strconv.Itoa(dataInt)
					}

				} else if dataSize == 2 {
					if eepromBytes[dataAddr] == 0xff {
						eepromResp.EepromData[i].Permission = 0
						eepromResp.EepromData[i].CurrValue = ""
					} else {
						dataInt := binary.LittleEndian.Uint16(eepromBytes[dataAddr:(dataAddr + 2)])
						eepromResp.EepromData[i].CurrValue = strconv.Itoa(int(dataInt))
					}

				} else if dataSize == 4 {
					if eepromBytes[dataAddr] == 0xff && eepromResp.EepromData[i].SubType != "ip" {
						eepromResp.EepromData[i].Permission = 0
						eepromResp.EepromData[i].CurrValue = ""
					} else {
						if dataSubType == "ip" {
							dataInt1 := int(eepromBytes[dataAddr])
							dataInt2 := int(eepromBytes[(dataAddr + 1)])
							dataInt3 := int(eepromBytes[(dataAddr + 2)])
							dataInt4 := int(eepromBytes[(dataAddr + 3)])
							eepromResp.EepromData[i].CurrValue = strconv.Itoa(dataInt1) + "." + strconv.Itoa(dataInt2) + "." + strconv.Itoa(dataInt3) + "." + strconv.Itoa(dataInt4)

						} else {
							dataInt := int(binary.LittleEndian.Uint32(eepromBytes[dataAddr:(dataAddr + 4)]))
							eepromResp.EepromData[i].CurrValue = strconv.Itoa(dataInt)
						}
					}
				} else if dataSize == 8 {
					if eepromBytes[dataAddr] == 0xff {
						eepromResp.EepromData[i].Permission = 0
						eepromResp.EepromData[i].CurrValue = ""
					} else {
						dataInt := int(binary.LittleEndian.Uint64(eepromBytes[dataAddr:(dataAddr + 8)]))
						eepromResp.EepromData[i].CurrValue = strconv.Itoa(dataInt)
					}
				}

			} else if dataType == "double" {

				if dataSize == 8 {
					dataInt64 := int64(binary.LittleEndian.Uint64(eepromBytes[dataAddr:(dataAddr + 8)]))
					dataDouble := math.Float64frombits(uint64(dataInt64))
					fmt.Printf("Converted data value: %f\n", dataDouble)
					fmt.Printf("8 Data 结果: %f\n", dataDouble)
					eepromResp.EepromData[i].CurrValue = strconv.FormatFloat(dataDouble, 'f', 4, 64)
				} else if dataSize == 4 {
					dataUint32 := (binary.LittleEndian.Uint32(eepromBytes[dataAddr:(dataAddr + 4)]))
					fmt.Printf("4 Data in hexadecimal: %x\n", dataUint32)
					dataDouble := math.Float32frombits(dataUint32)
					eepromResp.EepromData[i].CurrValue = strconv.FormatFloat(float64(dataDouble), 'f', 2, 32)
				}

			} else if dataType == "string" {
				if eepromBytes[dataAddr] == 0xff {
					eepromResp.EepromData[i].Permission = 0
					eepromResp.EepromData[i].CurrValue = ""
				} else {
					dataStr := string(eepromBytes[dataAddr:(dataAddr + dataSize)])
					eepromResp.EepromData[i].CurrValue = dataStr
				}

			}

		}

	}

	return eepromResp, true

}

type SerialOutputInfo struct {
	Id   [1]byte // 序号
	Data []byte  // 内容
}

func initSerialOutputInfo() []SerialOutputInfo {
	return []SerialOutputInfo{
		{[1]byte{1}, nil},
		{[1]byte{2}, nil},
		{[1]byte{3}, nil},
		{[1]byte{4}, nil},
		{[1]byte{5}, nil},
		{[1]byte{6}, nil},
	}

}

func ReqSetOutputFmt(c *Scale, req SRequest) (*ScaleRespMsg, error) {
	//先获取秤上的地址
	reg, err, res := openFactory(c)
	if err != nil || !res {
		return reg, err
	}
	composer := c.composer
	fn := composer.ComposeCmd

	reqMsg, _ := excuteSimpCmd(c, m.CMD_GET_SCALE_INFO, m.GET_SCALE_INFO_RESP)
	var scaleInfo SIFromScale
	serialOutputAddr := 0
	serialOutputMaxLenth := 0
	strData := reqMsg.MsgBody
	if str, ok := strData.(string); ok {
		if err := json.UnmarshalFromString(str, &scaleInfo); err != nil {
			l.Log.Error(err)
			return &ScaleRespMsg{}, fmt.Errorf("get serial format address fail")
		}
	} else {
		return &ScaleRespMsg{}, fmt.Errorf("get serial format address fail")
	}
	siAddrInfos := scaleInfo.AddrInfos
	fmt.Println(siAddrInfos)
	for _, jsonStr := range siAddrInfos {
		addrInfo, err := parseJSON(jsonStr)
		if err != nil {
			return &ScaleRespMsg{}, fmt.Errorf("get serial format address fail")
		}
		if addrInfo.Type == SI_SERIAL_OUTPUT {
			serialOutputAddr = addrInfo.Addr
			serialOutputMaxLenth = addrInfo.Lenth
		}
	}
	if serialOutputAddr == 0 || serialOutputMaxLenth == 0 {
		return &ScaleRespMsg{}, fmt.Errorf("get serial format address fail")
	}
	//开始分析json
	if req.ReqData == "" {
		return &ScaleRespMsg{}, nil
	}
	recvFileNames := req.ReqData
	fileDataInfoList := initSerialOutputInfo()
	//1 解出路径
	var fileList ReqSerialFileList
	err = json.Unmarshal([]byte(recvFileNames), &fileList)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
	}
	//2 根据路径分别作出完整的输出格式
	for i := 0; i < len(fileList.Paths); i++ {
		var fileNameStr = fileList.Paths[i]
		if len(fileNameStr) > 0 {
			fileId := int(fileNameStr[0] - '0')
			filePathStr := fileNameStr[1:]
			if csvFmtContent, err := os.ReadFile(filePathStr); err != nil {
				return &ScaleRespMsg{}, err
			} else {
				if bufferData, res := ParserSerialOutputFile(string(csvFmtContent)); res == false {
					return &ScaleRespMsg{}, err
				} else {
					fileDataInfoList[fileId-1].Data = bufferData.Bytes()
				}
			}
		}
	}
	//3 删除当前程序下的output.bin
	if err := deleteFile("output.bin"); err != nil {
		return &ScaleRespMsg{}, err
	}
	//4 将所有的格式合并为一个bin
	var totalOutputBuffer bytes.Buffer
	outputAddr := 0
	outputAddr = outputAddr + 12 //12个字节是开头地址的位置
	//json数据不为空则填实际地址，为空则补2个字节的0
	for i := 0; i < len(fileDataInfoList); i++ {
		lengthBytes := make([]byte, 2)
		if fileDataInfoList[i].Data != nil {
			binary.LittleEndian.PutUint16(lengthBytes, uint16(outputAddr))
			totalOutputBuffer.Write(lengthBytes)
			outputAddr = outputAddr + len(fileDataInfoList[i].Data)
		} else {
			binary.LittleEndian.PutUint16(lengthBytes, uint16(0))
			totalOutputBuffer.Write(lengthBytes)
		}
	}
	for i := 0; i < 6; i++ {
		if fileDataInfoList[i].Data != nil {
			totalOutputBuffer.Write(fileDataInfoList[i].Data)
		}
	}
	WriteDataToBin(totalOutputBuffer)

	if true { //TODO:
		// 读取bin文件
		data, err := os.ReadFile("output.bin")
		if err != nil {
			l.Log.Fatal(err)
		}

		// 擦除原本秤上的打印格式
		l.Log.Debug("erase flash on scale")
		addr := serialOutputAddr
		size := serialOutputMaxLenth
		loopCnt := size / 2048
		addrInLoop := addr
		for i := 0; i < loopCnt; i++ {
			cmd, timeoutMs, err := composer.ComposeCmd(composer, m.CMD_ERASE_FLASH, m.CmdData{Type: m.DATA_TYPE_INT, Data: addr})
			if err != nil {
				return &ScaleRespMsg{}, err
			}

			if res, err := perfCmdNwaitResult(c, cmd, m.ERASE_FLASH_RESP, timeoutMs); err != nil {
				return &ScaleRespMsg{}, err
			} else if res.MsgBody != "ok" {
				return &ScaleRespMsg{}, fmt.Errorf("erase fail")
			}
			addrInLoop += 2048
		}

		// 计算数据包数量
		packetCount := len(data) / DATA_LENGTH_256_TMAX
		if len(data)%DATA_LENGTH_256_TMAX != 0 {
			packetCount += 1
		}
		// 遍历所有数据包
		l.Log.Debug("send data package to scale")
		for i := 0; i < packetCount; i++ {
			// 计算本包数据
			start := i * DATA_LENGTH_256_TMAX
			end := start + DATA_LENGTH_256_TMAX
			if end > len(data) {
				end = len(data)
			}
			packetData := data[start:end]
			// 构建数据包
			// dataPackCmd := buildSendDataPacket(addr, packetData)
			packDataHexStr := hex.EncodeToString(packetData)
			cmd, timeoutMs, err := fn(composer, m.CMD_WRITE_FLASH_256, m.CmdData{Type: m.DATA_TYPE_STR, Data: fmt.Sprintf("%08x:%s", addr, packDataHexStr)})
			if err != nil {
				return &ScaleRespMsg{}, err
			}
			// 发送数据包
			if res, err := perfCmdNwaitResult(c, cmd, m.WRITE_DATA_FLASH_RESP, timeoutMs); err != nil {
				return &ScaleRespMsg{}, err
			} else if res.MsgBody != "ok" {
				return &ScaleRespMsg{}, fmt.Errorf("enable factory mode fail")
			}
			// 地址自增
			addr += 0x100
		}
		l.Log.Info("send bin ok")
	}
	return &ScaleRespMsg{m.SET_OUTPUT_FMT_RESP, "ok", c.Id}, nil
}

func retreiveRespMsgT2200(scaleId int64, data []byte) (*ScaleRespMsg, error) {
	// checkHead, get msgid, get msgtype, check if return code is 0x06, for success
	var err error
	respMsg := &ScaleRespMsg{MsgType: GlastWantRespMsgType, MsgBody: "", ScaleId: scaleId}
	if len(data) == 1 {
		if data[0] == 0x06 {
			respMsg.MsgBody = "ok"
			err = nil
		} else if data[0] == 0x15 {
			respMsg.MsgBody = "fail"
			err = nil
		} else {
			return respMsg, fmt.Errorf("unknown response")
		}
	} else { // weight data, same as C51 scale
		var msg WeightMsg
		if msg, err = retreiveWeightC51(data); err == nil {
			respMsg.MsgType = m.WEIGHT_DATA
			respMsg.MsgBody = msg
		}
	}
	return respMsg, err
}

func sendMsgIntoChsOrWeightToClient(s *Scale, msg *ScaleRespMsg) {
	if (*msg == ScaleRespMsg{}) {
		l.Log.Errorf("extractMsgTmaxScale error")
		return
	}

	chs := s.respChansMap[msg.MsgType]
	for _, ch := range chs {
		if len(ch) == 0 { // to avoid blocking, this kind of channel should only be used once
			ch <- msg
		}
	}
	if msg.MsgType == m.WEIGHT_DATA && s.isSendUnolicitedData { // skip sending weight data to client if it doesn't not register this message
		sendRespMsgClient(s, msg)
		return
	}
	if msg.MsgType == m.SCALE_PASSTH_DATA && s.isScalePassth { // skip sending weight data to client if it doesn't not register this message
		sendRespMsgClient(s, msg)
		return
	}
}

func sendRespMsgClient(s *Scale, msg *ScaleRespMsg) {
	if (*msg == ScaleRespMsg{}) {
		l.Log.Warnf("msg is ScaleRespMsg{}")
		return
	}

	if len(s.fromScaleMsgCh) < RECV_CH_SIZE { // no use in this moment
		if msgStr, err := json.MarshalToString(msg); err == nil {
			if s.client == nil {
				l.Log.Errorf("s.client is nil")
				return
			}

			l.Log.Tracef("%%%%%%%%%%%%%%: " + msgStr)
			s.client.sendCh <- []byte(msgStr)
		} else {
			l.Log.Errorf("marshal msg err: %v", err.Error())
		}
	} else {
		l.Log.Error("fromScaleMsgCh full")
	}
	return
}
