package svc

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"tmaxsrv/cmd"
	mcmd "tmaxsrv/cmd"
	m "tmaxsrv/comm"
	l "tmaxsrv/log"
	"tmaxsrv/util"
)

func (c *Scale) UpdateFirmware(name string) (*ScaleRespMsg, error) {
	EnFacMode(c)
	Reboot(c)
	pickerFn := c.MySerial.pickerFn
	c.MySerial.Close()

	output := make(chan string)
	done := make(chan error)
	go util.RunCommand(output, done, "./BootCommander.exe", "-t=xcp_rs232", "-d="+c.Pcnf.DevPath, "-b=57600", name)
	var err error
	isFinish := false
	isStartUpdate := false
	var cmdOutput string
	var percentage float32 = 0.0
	var updateTimeMs float32 = 0.0
	var respMsg *ScaleRespMsg
	respOk := &ScaleRespMsg{MsgType: m.UPDATE_FIRMWARE_RESP, MsgBody: "ok", ScaleId: c.Id}
	respFail := &ScaleRespMsg{MsgType: m.UPDATE_FIRMWARE_RESP, MsgBody: "fail", ScaleId: c.Id}
	for {
		if isFinish {
			break
		}
		select {
		case line := <-output:
			cmdOutput += line
			fmt.Println(line) // Print each line of output as it is received
			// send progress notification to UI
			if isStartUpdate {
				// check percentage and send progress notification to UI
				// percentage := getPercentage(cmdOutput)

				// respMsg := ScaleRespMsg{MsgType: m.UPDATE_FIRMWARE_PROGRESS, MsgBody: percentage, ScaleId: c.Id}
				// result, _ := json.Marshal(respMsg)
				// c.client.sendCh <- result
			}
			if !isStartUpdate && strings.Contains(cmdOutput, "Erasing") {
				isStartUpdate = true
				// inform UI update firmware is
				re := regexp.MustCompile(`Erasing (\d+) bytes`)
				match := re.FindStringSubmatch(cmdOutput)
				if len(match) > 1 {
					number := match[1]
					updateTimeInt, _ := strconv.Atoi(number)
					updateTimeMs = float32(updateTimeInt) / 2.6
					fmt.Println(number) // 输出: 64980
				}

				respMsg := ScaleRespMsg{MsgType: m.UPDATE_FIRMWARE_PROGRESS, MsgBody: "started", ScaleId: c.Id}
				result, _ := json.Marshal(respMsg)
				c.client.sendCh <- result
			}
		case err = <-done:
			if strings.Contains(cmdOutput, "Finishing programming session...[OK]") {
				respMsg100 := ScaleRespMsg{MsgType: m.UPDATE_FIRMWARE_PROGRESS, MsgBody: strconv.Itoa(100), ScaleId: c.Id}
				result, _ := json.Marshal(respMsg100)
				c.client.sendCh <- result
				time.Sleep(500 * time.Millisecond)
				respMsg = respOk
			} else {
				respMsg = respFail
			}
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("Command completed")
			}
			isFinish = true
		case <-time.After(time.Duration(250) * time.Millisecond):
			if isStartUpdate {
				percentage = percentage + 25000.0/updateTimeMs
				percentageInt := int(percentage)
				respMsg := ScaleRespMsg{MsgType: m.UPDATE_FIRMWARE_PROGRESS, MsgBody: strconv.Itoa(percentageInt), ScaleId: c.Id}
				result, _ := json.Marshal(respMsg)
				c.client.sendCh <- result
			}
		}
	}
	if c.MySerial, err = NewSerial(c.Pcnf, pickerFn); err != nil {
		l.Log.Error(err.Error())
	}
	return respMsg, nil
}

func (c *Scale) CheckSerialPort() (*ScaleRespMsg, error) {
	l.Log.Debug("check serial port")
	reqMsg, _ := excuteSimpCmd(c, m.CMD_EN_FAC_MODE, m.EN_FAC_MODE_RESP)
	if reqMsg.MsgBody == "ok" {
		return &ScaleRespMsg{MsgType: m.CHECK_SERIAL_PORT_RESP, MsgBody: "ok", ScaleId: c.Id}, nil

	}
	return &ScaleRespMsg{MsgType: m.CHECK_SERIAL_PORT_RESP, MsgBody: "fail", ScaleId: c.Id}, nil
}

func (c *Scale) GetBuildInfo() (*ScaleRespMsg, error) {
	l.Log.Debug("get build info")
	reqMsg, err := excuteSimpCmd(c, m.CMD_GET_BUILD_INFO, m.GET_BUILD_INFO_RESP)
	return reqMsg, err
}

func getPercentage(data string) string {
	percentage := "0%"
	pattern := `\[ *(\d+)%\][^[]*$`
	r := regexp.MustCompile(pattern)
	match := r.FindStringSubmatch(data)
	if len(match) > 1 {
		fmt.Println(match[1])
		percentage = match[1]
	} else {
		fmt.Println("No match found.")
	}

	return percentage + "%"
}

func (c *Scale) PerfZero() (*ScaleRespMsg, error) {
	l.Log.Debug("perform zero")
	_, err := EnFacMode(c)
	if err != nil {
		return &ScaleRespMsg{}, err //FLF
	}
	return excuteSimpCmd(c, m.CMD_ZERO, m.ZERO_CMD_RESP)
}

func (c *Scale) PerfTare() (*ScaleRespMsg, error) {
	l.Log.Debug("perform tare")
	_, err := EnFacMode(c)
	if err != nil {
		return &ScaleRespMsg{}, err //FLF
	}
	return excuteSimpCmd(c, m.CMD_TARE, m.TARE_CMD_RESP)
}

func (c *Scale) ReadWeight() (*ScaleRespMsg, error) {
	// if c.isOldC51Scale {
	//     return perfCmd(c, []byte(GET_WEIGHT_CMD))
	// } else {
	//     _, err := perfCmdNwaitResult(c, TMAX_READ_WEIGHT_CMD, WEIGHT_DATA_RESP)
	//     return err == nil
	// }
	return nil, nil // FIXME:
}

func (c *Scale) RegWeightData() (*ScaleRespMsg, error) { //FLF
	l.Log.Debug("register weight data")
	c.isSendUnolicitedData = true
	_, err := EnFacMode(c)
	if err != nil {
		return &ScaleRespMsg{}, err //FLF
	}
	// DisFacMode(c) // TODO: check return value
	// enable scale sending weighing info continually
	return excuteSimpCmd(c, m.CMD_EN_CONTINUE_MODE, m.REG_WEIGHT_RESP)
}

// func sendErrMsg(c *Scale, msg *ScaleRespMsg) {
//     msgStr, err := json.MarshalToString(msg)
//     if err != nil {
//         if c.client != nil {
//             c.client.sendCh <- []byte(msgStr)
//         }
//     }
// }

func (c *Scale) UnRegWeightData() (*ScaleRespMsg, error) {
	c.isSendUnolicitedData = false
	_, err := EnFacMode(c)
	if err != nil {
		return &ScaleRespMsg{}, err
	}
	// enable scale sending weighing info continually
	msg, err := perfCmdNwaitResult(c, cmd.DIS_CONT_MODE_CMD_TMAX, m.UNREG_WEIGHT_RESP)
	//sendErrMsg(c, msg)
	_, _ = EnFacMode(c)
	return msg, err
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

// 重启
func Reboot(s *Scale) (*ScaleRespMsg, error) {
	l.Log.Debug("send reboot cmd to scale")
	return excuteSimpCmd(s, m.CMD_REBOOT, m.UNKNOWN_DATA)
}

// 打开工厂模式
func EnFacMode(s *Scale) (*ScaleRespMsg, error) {
	l.Log.Debug("send enable factory mode cmd to scale")
	return excuteSimpCmd(s, m.CMD_EN_FAC_MODE, m.EN_FAC_MODE_RESP)
}

// 关闭工厂模式
func DisFacMode(s *Scale) (*ScaleRespMsg, error) {
	l.Log.Debug("send disable factory mode cmd to scale")
	return excuteSimpCmd(s, m.CMD_DIS_FAC_MODE, m.DIS_FAC_MODE_RESP)
}

func excuteSimpCmd(s *Scale, cmdType m.CmdType, respType m.RespMsgType) (*ScaleRespMsg, error) {
	scaleCmdExtractorFn := s.composer.ComposeCmd
	cmd, timeoutMs, err := scaleCmdExtractorFn(s.composer, cmdType, m.CmdData{})
	if err != nil {
		return &ScaleRespMsg{}, err
	}

	if res, err := perfCmdNwaitResult(s, cmd, respType, timeoutMs); err != nil {
		return &ScaleRespMsg{}, err
	} else {
		return res, nil
	}
}

// 打开BT透传模式
func EnPassthrough(s *Scale) (*ScaleRespMsg, error) {
	l.Log.Debug("send enable passthrough mode cmd to scale")
	return excuteSimpCmd(s, m.CMD_EN_PASSTH, m.EN_PASSTH_MODE_RESP)
}

// 关闭BT透传模式
func DisPassthrough(s *Scale) (*ScaleRespMsg, error) {
	return excuteSimpCmd(s, m.CMD_DIS_PASSTH, m.DIS_PASSTH_MODE_RESP)
}

var GExpectBTResp m.RespMsgType

func (c *Scale) ModifyBTName(name string) (*ScaleRespMsg, error) {
	cmd, timeoutMs, err := c.composer.ComposeCmd(c.composer, m.CMD_MODIFY_BT_NAME, m.CmdData{Type: m.DATA_TYPE_STR, Data: name})
	if err != nil {
		return &ScaleRespMsg{}, err
	}
	GExpectBTResp = m.MODIFY_BT_NAME_RESP
	return perfCmdNwaitResult(c, cmd, m.MODIFY_BT_NAME_RESP, timeoutMs)
}

var GExpectWifiResp m.RespMsgType

// Get AP list
func GetApList(c *Scale) (*ScaleRespMsg, error) {
	GExpectWifiResp = m.GET_AP_LIST_RESP
	return excuteSimpCmd(c, m.CMD_WIFI_GET_AP_LIST, m.GET_AP_LIST_RESP)
}

// Get Wifi AP info
func GetWifiApInfo(c *Scale) (*ScaleRespMsg, error) {
	GExpectWifiResp = m.GET_WIFI_AP_INFO_RESP
	return excuteSimpCmd(c, m.CMD_WIFI_GET_AP_INFO, m.GET_WIFI_AP_INFO_RESP)
}

// Send data to BT
func SendDataToBT(c *Scale, data string) (*ScaleRespMsg, error) {
	l.Log.Debug("Send data to BT")
	cmd, timeoutMs, err := c.composer.ComposeCmd(c.composer, m.CMD_BT_DATA_PASSTH, m.CmdData{Type: m.DATA_TYPE_STR, Data: data})
	if err != nil {
		return &ScaleRespMsg{}, err
	}
	GExpectBTResp = m.BT_PASSTH_DATA_RESP
	return perfCmdNwaitResult(c, cmd, m.BT_PASSTH_DATA_RESP, timeoutMs)
}

// Send data to Wifi
func SendDataToWifi(s *Scale, data string) (*ScaleRespMsg, error) {
	l.Log.Debug("Send data to WIFI")
	cmd, timeoutMs, err := s.composer.ComposeCmd(s.composer, m.CMD_WIFI_DATA_PASSTH, m.CmdData{Type: m.DATA_TYPE_STR, Data: data})
	if err != nil {
		return &ScaleRespMsg{}, err
	}
	GExpectBTResp = m.WIFI_PASSTH_DATA_RESP
	return perfCmdNwaitResult(s, cmd, m.WIFI_PASSTH_DATA_RESP, timeoutMs)
}

func SetWifiDynamicIp(s *Scale) (*ScaleRespMsg, error) {
	l.Log.Debug("set wifi to dynamic IP")
	GExpectWifiResp = m.SET_WIFI_DYNAMIC_IP_RESP
	return excuteSimpCmd(s, m.CMD_WIFI_EN_DHCP, m.SET_WIFI_DYNAMIC_IP_RESP)
}

func SetWifiStaticIp(s *Scale, ip string, gateway string, netmask string) (*ScaleRespMsg, error) {
	l.Log.Debug("set wifi to static IP")
	cmd, timeoutMs, err := s.composer.ComposeCmd(s.composer, m.CMD_WIFI_SET_STATIC_IP, m.CmdData{Type: m.DATA_TYPE_STR, Data: fmt.Sprintf("%s,%s,%s", ip, gateway, netmask)})
	if err != nil {
		return &ScaleRespMsg{}, err
	}
	GExpectWifiResp = m.SET_WIFI_STATIC_IP_RESP
	return perfCmdNwaitResult(s, cmd, m.SET_WIFI_STATIC_IP_RESP, timeoutMs)
}

// Connect to specifi AP
func ConnectWifiAp(s *Scale, ssid string, bssid string, passwd string) (*ScaleRespMsg, error) {
	l.Log.Debug("Connect to Wifi AP")
	cmd, timeoutMs, err := s.composer.ComposeCmd(s.composer, m.CMD_WIFI_CONN_AP, m.CmdData{Type: m.DATA_TYPE_STR, Data: fmt.Sprintf("%s,%s,%s", ssid, bssid, passwd)})
	if err != nil {
		return &ScaleRespMsg{}, err
	}
	GExpectWifiResp = m.CONNECT_AP_RESP
	return perfCmdNwaitResult(s, cmd, m.CONNECT_AP_RESP, timeoutMs)
}

// Get IP info from scale
func GetIpInfo(s *Scale) (*ScaleRespMsg, error) {
	l.Log.Debug("Get IP info from Scale")
	GExpectWifiResp = m.GET_IP_INFO_RESP
	return excuteSimpCmd(s, m.CMD_WIFI_GET_IP_INFO, m.GET_IP_INFO_RESP)
}

func GetIpMode(s *Scale) (*ScaleRespMsg, error) {
	l.Log.Debug("Get IP mode from Scale")
	GExpectWifiResp = m.GET_IP_MODE_RESP
	return excuteSimpCmd(s, m.CMD_WIFI_GET_IP_MODE, m.GET_IP_MODE_RESP)
}

func perfCmd(c *Scale, cmd []byte) bool {
	if err := writeScale(c, cmd); err != nil {
		l.Log.Error(err.Error())
		return false
	}
	return true
}

func perfCmdNwaitResult(c *Scale, cmd []byte, waitMsgType m.RespMsgType, timeoutMs ...int) (*ScaleRespMsg, error) {
	curTimeoutMs := 3000 // 3000 ms
	if len(timeoutMs) > 0 {
		curTimeoutMs = timeoutMs[0]
	}

	if curTimeoutMs == mcmd.CMD_TIMEOUT_IMMEDIATE {
		if err := writeScale(c, cmd); err != nil {
			l.Log.Error(err.Error())
			return &ScaleRespMsg{MsgType: waitMsgType, ScaleId: c.Id, MsgBody: "error"}, err
		}
		return &ScaleRespMsg{MsgType: waitMsgType, ScaleId: c.Id, MsgBody: "done"}, nil
	}

	ch := make(chan *ScaleRespMsg, 10)
	c.RegisterNotif(waitMsgType, ch)
	defer func() {
		c.UnRegisterNotif(waitMsgType, ch)
		c.isWaintingResp = false
	}()

	GlastWantRespMsgType = waitMsgType
	c.isWaintingResp = true
	if err := writeScale(c, cmd); err != nil {
		l.Log.Error(err.Error())
		return &ScaleRespMsg{MsgType: waitMsgType, ScaleId: c.Id, MsgBody: "error"}, err
	}

	var ret *ScaleRespMsg
	fmt.Printf("================wait: %v\n", waitMsgType)
	select {
	case ret = <-ch:
	case <-time.After(time.Duration(curTimeoutMs) * time.Millisecond):
		ret = &ScaleRespMsg{MsgType: waitMsgType, ScaleId: c.Id, MsgBody: "timeout"}
		return ret, fmt.Errorf("no response, time out")
	}
	fmt.Printf("^^^^^^^^^^^^^^^^Got: %v\n", waitMsgType)
	return ret, nil
}

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

func GetToScaleCmd(scaleCat m.ScaleCat, cmdType SReqType) []byte {
	switch scaleCat {

	}

	return nil
}

func Req2CmdForC51(req SReqType) ([]byte, error) {
	fmt.Println("Req2CmdForC51")
	return nil, nil
}

// // 打开工厂模式
// func enFacModeCmdT2200() []byte {
//     // 构建包头
//     packet := make([]byte, OPEN_FAC_CHUNK_SIZE)
//     binary.BigEndian.PutUint16(packet[0:2], PACKET_HEAD)

//     // 构建命令ID与命令类型
//     packet[2] = 0x05
//     packet[3] = 0xF1

//     // 添加包尾
//     binary.BigEndian.PutUint16(packet[OPEN_FAC_CHUNK_SIZE-2:], PACKET_TAIL)

//     return packet
// }

// // 打开工厂模式
// func disFacModeCmdT2200() []byte {
//     // 构建包头
//     packet := make([]byte, OPEN_FAC_CHUNK_SIZE)
//     binary.BigEndian.PutUint16(packet[0:2], PACKET_HEAD)

//     // 构建命令ID与命令类型
//     packet[2] = 0x05
//     packet[3] = 0xF2

//     // 添加包尾
//     binary.BigEndian.PutUint16(packet[OPEN_FAC_CHUNK_SIZE-2:], PACKET_TAIL)

//     return packet
// }

// // 擦除原本秤上的打印格式
// func eraseCmdT2200(addr uint32) []byte { // erase size will 2K
//     // 构建包头
//     packet := make([]byte, EARSE_CHUNK_SIZE)
//     binary.BigEndian.PutUint16(packet[0:2], PACKET_HEAD)
//     // 构建命令ID与命令类型
//     packet[2] = CMD_ERASE
//     packet[3] = CMD_FLASH

//     // 构建地址
//     binary.BigEndian.PutUint32(packet[4:8], addr)

//     // 擦除长度
//     binary.BigEndian.PutUint16(packet[8:10], CMD_ERASE_SIZE)

//     // 计算与添加校验码
//     checksum := utils.Crc32MPEG2(packet[2 : EARSE_CHUNK_SIZE-6])
//     binary.BigEndian.PutUint32(packet[EARSE_CHUNK_SIZE-6:], checksum)

//     // 添加包尾
//     binary.BigEndian.PutUint16(packet[EARSE_CHUNK_SIZE-2:], PACKET_TAIL)

//     return packet
// }

// func wrDataCmdT2200(addr uint32, data []byte) []byte {
//     // 构建包头
//     packet := make([]byte, FILE_CHUNK_SIZE)
//     binary.BigEndian.PutUint16(packet[0:2], PACKET_HEAD)

//     // 构建命令ID与命令类型
//     packet[2] = CMD_IDENTIFY
//     packet[3] = CMD_TYPE

//     // 构建地址
//     binary.BigEndian.PutUint32(packet[4:8], addr)

//     // 构建数据长度
//     binary.BigEndian.PutUint16(packet[8:10], uint16(len(data)))

//     // 复制数据
//     copy(packet[10:], data)

//     // 计算与添加校验码
//     checksum := utils.Crc32MPEG2(packet[2 : FILE_CHUNK_SIZE-6])
//     binary.BigEndian.PutUint32(packet[FILE_CHUNK_SIZE-6:], checksum)

//     // 添加包尾
//     binary.BigEndian.PutUint16(packet[FILE_CHUNK_SIZE-2:], PACKET_TAIL)

//     return packet
// }

// func Req2CmdForT2200(req SReqType, reqData string) ([]byte, error) {
//     fmt.Println("Req2CmdForT2200")
//     switch req {
//     // case SREQ_EN_FAC_MODE:
//     //     return enFacModeCmdT2200(), nil
//     // case SREQ_DIS_FAC_MODE:
//     //     return disFacModeCmdT2200(), nil
//     // case SREQ_ERASE_FLASH:
//     //     // TODO: reqData: format sequence 0-3
//     //     seqNo, err := strconv.Atoi(reqData)
//     //     if err != nil {
//     //         return nil, err
//     //     }
//     //     return eraseCmdT2200(uint32(T2200_PRN_FMT_BASE_ADDR + seqNo*CMD_ERASE_SIZE)), nil
//     // case SREQ_READ_FLASH:
//     //     return T2200_READ_FLASH, nil
//     // case SREQ_WRITE_FLASH:
//     //     seqNo, err := strconv.Atoi(reqData[:1])
//     //     if err != nil {
//     //         fmt.Println("Invalid sequence number")
//     //         return nil, err
//     //     }
//     //     offset, err := strconv.Atoi(reqData[2:8])
//     //     if err != nil {
//     //         fmt.Println("Invalid offset number")
//     //         return nil, err
//     //     }

//     //     // Extract byte array
//     //     data, err := hex.DecodeString(reqData[9:])
//     //     if err != nil {
//     //         fmt.Println("Invalid hex string")
//     //         return nil, err
//     //     }
//     // return wrDataCmdT2200(uint32(T2200_PRN_FMT_BASE_ADDR+seqNo*CMD_ERASE_SIZE+offset), data), nil
//     }

//     return nil, nil
// }

// // func Req2CmdForJWP(req SReqType) ([]byte, error) {
// //     fmt.Println("Req2CmdForJWP")
// //     return nil, nil
// // }

// // func Req2CmdForTMAX(req SReqType) ([]byte, error) {
// //     fmt.Println("Req2CmdForTMAX")
// //     return nil, nil
// // }

// 创建ScaleCat的函数映射
// var scaleCatFuncMap map[ScaleCat]func(req SReqType, reqData string) ([]byte, error)
var cmdComposerFuncMap map[m.ScaleCat]m.CmdComposer

// scaleCatFuncMap := make(map[ScaleCat]func(req SReqType, reqData string) ([]byte, error))
// // scaleCatFuncMap[SCALE_C51] = Req2CmdForC51
// scaleCatFuncMap[SCALE_T2200] = Req2CmdForT2200
// scaleCatFuncMap[SCALE_JWP] = Req2CmdForJWP
// scaleCatFuncMap[SCALE_TMAX] = Req2CmdForTMAX

// cmdComposerFuncMap := make(map[m.ScaleCat]func(req SReqType, reqData string) ([]byte, error))
// cmdComposerFuncMap[m.SCALE_T2200] = ComposerT2200

func init() {
	cmdComposerFuncMap = make(map[m.ScaleCat]m.CmdComposer)
	cmdComposerFuncMap[m.SCALE_T2200] = *cmd.NewComposerT2200()
	cmdComposerFuncMap[m.SCALE_TMAX] = *cmd.NewComposerTMAX()
}
