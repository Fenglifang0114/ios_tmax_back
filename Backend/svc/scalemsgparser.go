package svc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gitteamer/log"

	"tmaxsrv/cmd"
	m "tmaxsrv/comm"
	"tmaxsrv/util"
)

const (
	MODIFY_BT_OK_RESP string = "TTM:OK\r\n"
)

const (
	CONNECT_AP_OK_RESP          string = "\r\nOK\r\n"
	CONNECT_AP_FAIL_RESP        string = "+CWJAP:"
	SET_WIFI_DYNAMIC_IP_OK_RESP string = "\r\nOK\r\n"
	SET_WIFI_STATIC_IP_OK_RESP  string = "\r\nOK\r\n"
	GET_AP_INFO_OK_RESP         string = "\r\nOK\r\n"
	GET_IP_INFO_OK_RESP         string = "\r\nOK\r\n"
	GET_IP_MODE_OK_RESP         string = "\r\nOK\r\n"
	CHANG_WIFI_MODE_OK_RESP     string = "\r\nOK\r\n"
	CHANG_WIFI_MODE_FAIL_RESP   string = "\r\n+CWMODE:\r\n"
)

const (
	RSSI_MAX = -50  // maximum strength of signal in dBm
	RSSI_MIN = -100 // minimum strength of signal in dBm
)

// SI_  代表scale info 的缩写
const (
	SI_MODEL_NAME    = 1
	SI_SCALE_SN      = 2
	SI_DEF_PRN_INFO  = 3
	SI_FREE_PRN_INFO = 4
	SI_SERIAL_OUTPUT = 5
	SI_PLU_INFO      = 6
	SI_OL_INFO       = 7
	SI_UL_INFO       = 8
)

var responseHandlerMap map[m.RespMsgType]func(int64, []byte) (ScaleRespMsg, int)

func init() {
	util.CmdsRespMap = util.CmdMap{
		0xe107:                          m.WEIGHT_DATA,
		0xe103:                          m.ZERO_CMD_RESP,
		0xe105:                          m.TARE_CMD_RESP,
		0xe101:                          m.WEIGHT_DATA_RESP,
		0xfff3:                          m.REG_WEIGHT_RESP,
		0xe108:                          m.UNREG_WEIGHT_RESP,
		0xfff6:                          m.GET_RECS_RESP,
		0xfff7:                          m.ADD_REC_RESP,
		0xfff8:                          m.DEL_REC_RESP,
		cmd.CMDID_REBOOT_TMAX:           m.REBOOT_RESP,
		0x0556:                          m.GET_BUILD_INFO_RESP,
		cmd.CMDID_READ_SCALE_INFO_TMAX:  m.GET_SCALE_INFO_RESP,
		cmd.CMDID_GET_SCALE_TIME_TMAX:   m.GET_SCALE_TIME_RESP,
		0x05f1:                          m.EN_FAC_MODE_RESP,
		0x05f2:                          m.DIS_FAC_MODE_RESP,
		0x05f3:                          m.EN_PASSTH_MODE_RESP,
		0x05f4:                          m.DIS_PASSTH_MODE_RESP,
		cmd.CMDID_GET_FACTORY_INFO_TMAX: m.GET_FACTORY_INFO_RESP,
		cmd.CMDID_GET_RANDOM_DATA:       m.GET_RANDOM_DATA_RESP,

		cmd.CMDID_ERASE_FLASH_TMAX:       m.ERASE_FLASH_RESP, //FLF//
		cmd.CMDID_WRITE_FLASH_TMAX:       m.WRITE_DATA_FLASH_RESP,
		0xff11:                           m.DOWN_PRN_FMT_RESP,
		0xff12:                           m.ERR_SERIAL_RESP,
		0xff13:                           m.GET_AP_LIST_RESP,
		0xff14:                           m.RESCAN_AP_LIST_RESP,
		0xff15:                           m.SET_WIFI_DYNAMIC_IP_RESP,
		0xff16:                           m.SET_WIFI_STATIC_IP_RESP,
		0xff17:                           m.GET_IP_INFO_RESP,
		0xff18:                           m.MODIFY_BT_NAME_RESP,
		0xff19:                           m.NO_RESP,
		0xf201:                           m.BT_PASSTH_DATA_RESP,
		0xf202:                           m.WIFI_PASSTH_DATA_RESP,
		0xff22:                           m.PRT_PASSTH_DATA_RESP,
		cmd.CMDID_SCALE_PASSTH_DATA_TMAX: m.SCALE_PASSTH_DATA,
		cmd.CMDID_DOWN_PLU_TMAX:          m.DOWN_PLU_RESP,
		cmd.CMDID_DEL_PLU_TMAX:           m.DEL_PLU_RESP,
		cmd.CMDID_SET_SCALE_TIME_TMAX:    m.SET_SCALE_TIME_RESP,
		cmd.CMDID_INSERT_PLU_TMAX:        m.INSERT_PLU_ADDR_RESP,
		cmd.CMDID_READ_FLASH_TMAX:        m.READ_FLASH_DATA_RESP,
		cmd.CMDID_ERASE_INSERT_PLU_TMAX:  m.ERASE_INSERT_PLU_RESP,
		cmd.CMDID_MODIFY_VAR_TMAX:        m.MODIFY_VAR_RESP,
		cmd.CMDID_EN_FACTORY_MODE:        m.EN_FACTORY_MODE_RESP,

		0xff25: m.UNKNOWN_DATA,
	}

	responseHandlerMap = map[m.RespMsgType]func(int64, []byte) (ScaleRespMsg, int){
		m.WEIGHT_DATA:               handleWeightDataMsg,
		m.ZERO_CMD_RESP:             handleZeroCmdResp,
		m.TARE_CMD_RESP:             handleTareCmdResp,
		m.WEIGHT_DATA_RESP:          handleWeightDataResp,
		m.REG_WEIGHT_RESP:           handleRegWeightResp,
		m.UNREG_WEIGHT_RESP:         handleUnregWeightResp,
		m.GET_RECS_RESP:             handleGetRecsResp,
		m.ADD_REC_RESP:              handleAddRecResp,
		m.DEL_REC_RESP:              handleDelRecResp,
		m.EN_FAC_MODE_RESP:          handleEnFacModeResp,
		m.DIS_FAC_MODE_RESP:         handleDisFacModeResp,
		m.EN_PASSTH_MODE_RESP:       handleEnPassthModeResp,
		m.DIS_PASSTH_MODE_RESP:      handleDisPassthModeResp,
		m.ERASE_FLASH_RESP:          handleEraseFlashResp,
		m.WRITE_DATA_FLASH_RESP:     handleWriteDataFlashResp,
		m.DOWN_PRN_FMT_RESP:         handleDownPrnFmtResp,
		m.ERR_SERIAL_RESP:           handleErrSerialResp,
		m.GET_AP_LIST_RESP:          handleGetApListResp,
		m.RESCAN_AP_LIST_RESP:       handleRescanApListResp,
		m.SET_WIFI_DYNAMIC_IP_RESP:  handleSetWifiDynamicIpResp,
		m.SET_WIFI_STATIC_IP_RESP:   handleSetWifiStaticIpResp,
		m.GET_IP_INFO_RESP:          handleGetIpInfoResp,
		m.GET_IP_MODE_RESP:          handleGetIpModeResp,
		m.MODIFY_BT_NAME_RESP:       handleModifyBtNameResp,
		m.BT_PASSTH_DATA_RESP:       handleBTPassthResp,
		m.WIFI_PASSTH_DATA_RESP:     handleWifiPassthResp,
		m.GET_BUILD_INFO_RESP:       handleGetBuildInfoResp,
		m.GET_SCALE_INFO_RESP:       handleGetScaleInfoResp,
		m.GET_FACTORY_INFO_RESP:     handleGetFactoryInfoResp,
		m.GET_SCALE_TIME_RESP:       handleGetScaleTimeResp,
		m.SET_SCALE_TIME_RESP:       handleSetScaleTimeResp,
		m.DOWN_PLU_RESP:             handleDownPluResp,
		m.DEL_PLU_RESP:              handleDelPluResp,
		m.INSERT_PLU_ADDR_RESP:      handleInsertPluResp,
		m.READ_FLASH_DATA_RESP:      handleReadFlashDataResp,
		m.ERASE_INSERT_PLU_RESP:     handleEraseInsertPluResp,
		m.REBOOT_RESP:               handleRebootResp,
		m.MODIFY_VAR_RESP:           handleModifyVarResp,
		m.EN_FACTORY_MODE_RESP:      handleEnFactoryModeResp,
		m.GET_RANDOM_DATA_RESP:      handleGetRandomDataResp,
		m.DOWN_DEFAULT_PRN_FMT_RESP: handleDownDefaultPrnFmtResp,
	}

	// example usage: call the handler for the WEIGHT_DATA message
	// msg := "some message"
	// responseHandlerMap[WEIGHT_DATA](msg)
}

func extractMessageTMAX(scaleId int64, bufs *util.CircularBuffer, msgType m.RespMsgType) ScaleRespMsg {
	data := bufs.PeekAll()
	handler := responseHandlerMap[msgType]
	if handler == nil {
		log.Error("handler not found, msgType: %v", msgType)
	}
	resp, shouldRemoveLen := handler(scaleId, data)
	bufs.DequeueN(shouldRemoveLen)

	return resp
}

func extractScalePassthDataTMAX(s *Scale, bufs *util.CircularBuffer, msgType m.RespMsgType) ScaleRespMsg {
	data := bufs.PeekAll()
	resp, shouldRemoveLen := handleScalePassthData(s.Id, data, s.IsScalePassthHex)
	bufs.DequeueN(shouldRemoveLen)

	return resp
}

func handleWeightDataMsg(scaleId int64, data []byte) (ScaleRespMsg, int) {
	weightMsg, err := retrieveWeight(data)
	if err != nil {
		return ScaleRespMsg{}, len(data)
	}
	weightStr, err := json.MarshalToString(weightMsg)
	if err == nil {
		fmt.Printf("%v", weightStr)
	}
	respMsg := ScaleRespMsg{MsgType: m.WEIGHT_DATA, MsgBody: weightStr, ScaleId: scaleId}

	return respMsg, len(data)
}

func retrieveWeight(data []byte) (WeightMsg, error) {
	dataStr := string(data)
	dataStr = strings.TrimSpace(dataStr)
	fields := strings.Split(dataStr, ",")

	if len(fields) != 3 {
		if len(fields[0]) < MIN_PACK_SIZE {
			return WeightMsg{}, fmt.Errorf("no packet")
		}

		weightVal := strings.TrimRight(dataStr, "\r\n")
		weightVal = strings.TrimSpace(weightVal)
		if weightVal == "--OL--" || weightVal == "--UL--" {
			return WeightMsg{WeightVal: weightVal, WeightUnit: ""}, nil
		}

		// Regex pattern to match "ST,NT90PCS" or "ST,GS 80%"
		pattern := `^ZE,ST,\s*([A-Za-z0-9]+)\s*([%A-Za-z]+)$`
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(weightVal)
		if len(match) == 3 {
			weightVal := match[1]
			weightUnit := match[2]
			return WeightMsg{WeightVal: weightVal, WeightUnit: weightUnit}, nil
		}

		return WeightMsg{}, fmt.Errorf("invalid weight format: %v", weightVal)
	}

	weightMsg := WeightMsg{}
	weightMsg.IsZero = strings.Contains(strings.TrimSpace(fields[0]), "ZE")
	weightMsg.IsStable = strings.Contains(strings.TrimSpace(fields[1]), "ST")
	weightMsg.IsNet = strings.Contains(strings.TrimSpace(fields[2]), "NT")

	regexp, err := regexp.Compile(`([0-9:.-]+)\s*([a-zA-Z%:]+)`)
	if err != nil {
		return WeightMsg{}, err
	}

	match := regexp.FindStringSubmatch(strings.ReplaceAll(fields[2], " ", ""))
	if len(match) != 3 {
		return WeightMsg{}, fmt.Errorf("finding substring error: %v", fields[2])
	}

	weightMsg.WeightVal = strings.TrimSpace(match[1])
	weightMsg.WeightUnit = strings.TrimSpace(match[2])

	return weightMsg, nil
}

func handleZeroCmdResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	msg := ScaleRespMsg{}
	if data[0] == 0x06 {
		msg.MsgType = m.ZERO_CMD_RESP
		msg.ScaleId = scaleId
		msg.MsgBody = "ok"
	} else {
		msg.MsgType = m.ZERO_CMD_RESP
		msg.ScaleId = scaleId
		msg.MsgBody = "fail"
	}
	return msg, len(data)
}

func handleTareCmdResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	msg := ScaleRespMsg{}
	if data[0] == 0x06 {
		msg.MsgType = m.TARE_CMD_RESP
		msg.ScaleId = scaleId
		msg.MsgBody = "ok"
	} else {
		msg.MsgType = m.TARE_CMD_RESP
		msg.ScaleId = scaleId
		msg.MsgBody = "fail"
	}
	return msg, len(data)
}

func handleWeightDataResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	// TODO: Implement function
	return ScaleRespMsg{}, 0
}

func handleRegWeightResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	msg := ScaleRespMsg{}
	if data[0] == 0x06 {
		msg.MsgType = m.REG_WEIGHT_RESP
		msg.ScaleId = scaleId
		msg.MsgBody = "ok"
	} else {
		msg.MsgType = m.REG_WEIGHT_RESP
		msg.ScaleId = scaleId
		msg.MsgBody = "fail"
	}
	return msg, len(data)
}

func handleUnregWeightResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	msg := ScaleRespMsg{}
	if data[0] == 0x06 {
		msg.MsgType = m.UNREG_WEIGHT_RESP
		msg.ScaleId = scaleId
		msg.MsgBody = "ok"
	} else {
		msg.MsgType = m.UNREG_WEIGHT_RESP
		msg.ScaleId = scaleId
		msg.MsgBody = "fail"
	}
	return msg, len(data)
}

func handleGetBuildInfoResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	msg := ScaleRespMsg{}
	// 5A A5 00 30 05 56 00 78 78 78 78 78 78 78 78 2D 78 78 78 78 2D 78 78 78 78 2D 78 78 78 78 2D 78 78 78 78 78 78 78 78 78 78 78 78 00 1C EE D5 0B A5 5A 5A A5 00 0C 05 56 00 06 8E 0B A1 55 A5 5A
	str := string(data)

	if len(str) >= 36 {
		msg.MsgType = m.GET_BUILD_INFO_RESP
		msg.ScaleId = scaleId
		msg.MsgBody = str[:36]
	} else if data[0] == 0x06 {

	} else {
		msg.MsgType = m.GET_BUILD_INFO_RESP
		msg.ScaleId = scaleId
		msg.MsgBody = "fail"
	}
	return msg, len(data)
}

type SIFromScale struct {
	ScaleSn   string   `json:"ScaleSn"`
	ModelName string   `json:"ModelName"`
	AddrInfos []string `json:"AddrInfos"`
}

type SIAddrInfos struct {
	Type     int `json:"Type"`
	Addr     int `json:"Addr"`
	Lenth    int `json:"Lenth"`
	EraseLen int `json:"EraseLen"`
}

type FIFromScale struct {
	ScaleSn   string `json:"ScaleSn"`
	ModelName string `json:"ModelName"`
}

func handleSetScaleTimeResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	if data[0] == 0x06 {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.SET_SCALE_TIME_RESP, MsgBody: "ok"}, len(data)
	} else {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.SET_SCALE_TIME_RESP, MsgBody: "fail"}, len(data)
	}
	// TODO: Implement function
}

func handleGetScaleInfoResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	//TODO:       20231101@FLF
	//此处处理收到的数据
	var scaleInfo SIFromScale
	var siAddrInfos SIAddrInfos
	msg := ScaleRespMsg{}
	// 0 75 1 5 84 45 77 65 88 2 10 84 45 83 67 65 76 69 48 48 49 3 12 8 0 48 0 0 0 32 0 0 0 8 0 4 12 8 1 224 0 0 0 32 0 0 0 8 0 5 12 8 1 216 0 0 0 8 0 0 0 8 0 6 12 0 6 160 0 0 16 0 0 0 0 16 0
	if data[0] == 0x15 || data[0] == 0x06 {
		msg.MsgType = m.GET_SCALE_INFO_RESP
		msg.ScaleId = scaleId
		msg.MsgBody = "fail"
		return msg, len(data)

	}
	dataLen := int(binary.BigEndian.Uint16(data[:2]))
	var byteLen int
	fmt.Println(dataLen)
	for i := 2; i < dataLen; i++ {
		println(data[i])
		switch data[i] {
		case SI_MODEL_NAME:
			byteLen = int(data[i+1])
			scaleInfo.ModelName = string(data[i+2 : i+2+byteLen])
			i = i + 1 + byteLen
		case SI_SCALE_SN:
			byteLen = int(data[i+1])
			scaleInfo.ScaleSn = string(data[i+2 : i+2+byteLen])
			i = i + 1 + byteLen
		case SI_DEF_PRN_INFO, SI_FREE_PRN_INFO, SI_SERIAL_OUTPUT, SI_PLU_INFO, SI_OL_INFO, SI_UL_INFO:
			byteLen = int(data[i+1])
			var infoByte = data[i+2 : i+2+byteLen]
			siAddrInfos.Type = int(data[i])
			siAddrInfos.Addr, siAddrInfos.Lenth, siAddrInfos.EraseLen = getInfoAddrLen(infoByte)
			addrInfoStr, _ := json.MarshalToString(siAddrInfos)
			scaleInfo.AddrInfos = append(scaleInfo.AddrInfos, addrInfoStr)
			i = i + 1 + byteLen
		default:

		}
	}

	msg.MsgType = m.GET_SCALE_INFO_RESP
	msg.ScaleId = scaleId
	msg.MsgBody, _ = json.MarshalToString(scaleInfo)

	return msg, len(data)
}

func handleGetFactoryInfoResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	//     20240417@FLF
	// 5A A5 00 1E 05 F6 00 05 54 2D 4D 41 58 0C 31 30 38 30 30 38 30 32 30 30 30 38 62 C6 A8 39 A5 5A
	var factoryInfo FIFromScale
	msg := ScaleRespMsg{}
	msg.MsgType = m.GET_FACTORY_INFO_RESP
	msg.ScaleId = scaleId
	msg.MsgBody = "fail"

	tmpInt := int(data[0])
	endIndex := tmpInt + 1
	if endIndex > len(data) {
		return msg, len(data)
	}
	factoryInfo.ModelName = string(data[1:endIndex])

	if endIndex+1 > len(data) {
		return msg, len(data)
	}

	tmpInt = int(data[endIndex])
	endIndex1 := tmpInt + 1 + endIndex
	if endIndex1 > len(data) {
		return msg, len(data)
	}
	factoryInfo.ScaleSn = string(data[endIndex+1 : endIndex1])

	msg.MsgBody, _ = json.MarshalToString(factoryInfo)

	return msg, len(data)
}

func handleGetScaleTimeResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	//TODO:       20231101@FLF
	//此处处理收到的数据
	msg := ScaleRespMsg{}

	if len(data) < 4 {
		msg.MsgType = m.GET_SCALE_TIME_RESP
		msg.ScaleId = scaleId
		msg.MsgBody = "fail"
		return msg, len(data)
	}
	dataTimeInt := int(binary.BigEndian.Uint32(data[0:4]))
	strTime := strconv.Itoa(dataTimeInt)
	if strTime == "" {
		msg.MsgType = m.GET_SCALE_TIME_RESP
		msg.ScaleId = scaleId
		msg.MsgBody = "fail"
		return msg, len(data)
	}

	msg.MsgType = m.GET_SCALE_TIME_RESP
	msg.ScaleId = scaleId
	msg.MsgBody = "ok" + "," + strTime

	return msg, len(data)
}

func getInfoAddrLen(data []byte) (int, int, int) {
	return int(binary.BigEndian.Uint32(data[:4])), int(binary.BigEndian.Uint32(data[4:8])), int(binary.BigEndian.Uint32(data[8:12]))

}

func handleGetRecsResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	// TODO: Implement function
	return ScaleRespMsg{}, 0
}

func handleAddRecResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	// TODO: Implement function
	return ScaleRespMsg{}, 0
}

func handleDelRecResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	// TODO: Implement function
	return ScaleRespMsg{}, 0
}

func handleEnFacModeResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	if data[0] == 0x06 {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.EN_FAC_MODE_RESP, MsgBody: "ok"}, len(data)
	} else {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.EN_FAC_MODE_RESP, MsgBody: "fail"}, len(data)
	}
}

// 最新修改的打开工厂模式
func handleEnFactoryModeResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	if data[0] == 0x06 {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.EN_FACTORY_MODE_RESP, MsgBody: "ok"}, len(data)
	} else {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.EN_FACTORY_MODE_RESP, MsgBody: "fail"}, len(data)
	}
}

func handleGetRandomDataResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	if len(data) == 2 {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.GET_RANDOM_DATA_RESP, MsgBody: data}, len(data)
	} else {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.GET_RANDOM_DATA_RESP, MsgBody: "fail"}, len(data)
	}
}

func handleRebootResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	if data[0] == 0x06 {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.REBOOT_RESP, MsgBody: "ok"}, len(data)
	} else {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.REBOOT_RESP, MsgBody: "fail"}, len(data)
	}
}

func handleDisFacModeResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	if data[0] == 0x06 {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.DIS_FAC_MODE_RESP, MsgBody: "ok"}, len(data)
	} else {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.DIS_FAC_MODE_RESP, MsgBody: "fail"}, len(data)
	}
}

func handleEnPassthModeResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	if data[0] == 0x06 {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.EN_PASSTH_MODE_RESP, MsgBody: "ok"}, len(data)
	} else {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.EN_PASSTH_MODE_RESP, MsgBody: "fail"}, len(data)
	}
}

func handleDisPassthModeResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	if data[0] == 0x06 {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.DIS_PASSTH_MODE_RESP, MsgBody: "ok"}, len(data)
	} else {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.DIS_PASSTH_MODE_RESP, MsgBody: "fail"}, len(data)
	}
}

func handleEraseFlashResp(scaleId int64, data []byte) (ScaleRespMsg, int) { //FLF
	if data[0] == 0x06 {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.ERASE_FLASH_RESP, MsgBody: "ok"}, len(data)
	} else {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.ERASE_FLASH_RESP, MsgBody: "fail"}, len(data)
	}
}

func handleWriteDataFlashResp(scaleId int64, data []byte) (ScaleRespMsg, int) { //FLF
	if data[0] == 0x06 {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.WRITE_DATA_FLASH_RESP, MsgBody: "ok"}, len(data)
	} else {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.WRITE_DATA_FLASH_RESP, MsgBody: "fail"}, len(data)
	}
}

func handleModifyVarResp(scaleId int64, data []byte) (ScaleRespMsg, int) { //FLF
	if data[0] == 0x06 {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.MODIFY_VAR_RESP, MsgBody: "ok"}, len(data)
	} else {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.MODIFY_VAR_RESP, MsgBody: "fail"}, len(data)
	}
}

func handleDownDefaultPrnFmtResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	// TODO: Implement function
	return ScaleRespMsg{}, 0
}

func handleDownPrnFmtResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	// TODO: Implement function
	return ScaleRespMsg{}, 0
}

func handleDownPluResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	// TODO: Implement function
	return ScaleRespMsg{}, 0
}

func handleGetAllEepromInfoResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	// TODO: Implement function
	return ScaleRespMsg{}, 0
}

func handleDelPluResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	if data[0] == 0x06 {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.DEL_PLU_RESP, MsgBody: "ok"}, len(data)
	} else {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.DEL_PLU_RESP, MsgBody: "fail"}, len(data)
	}
	// TODO: Implement function
}

func handleEraseInsertPluResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	if data[0] == 0x06 {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.ERASE_INSERT_PLU_RESP, MsgBody: "ok"}, len(data)
	} else {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.ERASE_INSERT_PLU_RESP, MsgBody: "fail"}, len(data)
	}

}

func handleInsertPluResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	msg := ScaleRespMsg{}
	str := string(data)
	if len(str) >= 12 {
		msg.MsgType = m.INSERT_PLU_ADDR_RESP
		msg.ScaleId = scaleId
		msg.MsgBody = str[:12]
	} else if data[0] == 0x06 {

	} else {
		msg.MsgType = m.INSERT_PLU_ADDR_RESP
		msg.ScaleId = scaleId
		msg.MsgBody = "fail"
	}
	return msg, len(data)
}

func handleReadFlashDataResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	msg := ScaleRespMsg{}
	str := string(data)
	if len(str) >= 0 {
		msg.MsgType = m.READ_FLASH_DATA_RESP
		msg.ScaleId = scaleId
		msg.MsgBody = str[:len(data)]
	} else if data[0] == 0x06 {

	} else {
		msg.MsgType = m.READ_FLASH_DATA_RESP
		msg.ScaleId = scaleId
		msg.MsgBody = "fail"
	}
	return msg, len(data)
}

func handleErrSerialResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	// TODO: Implement function
	return ScaleRespMsg{}, 0
}

func handleScalePassthData(scaleId int64, data []byte, isHexMode bool) (ScaleRespMsg, int) {
	// find \r\n
	target := []byte{0x0d, 0x0a}

	index := bytes.LastIndex(data, target)
	if index == -1 {
		return ScaleRespMsg{}, 0
	}
	var passthStr string
	if isHexMode {
		for _, b := range data[0 : index+2] {
			passthStr += fmt.Sprintf("%02X ", b)
		}

	} else {
		passthStr = string(data[0 : index+2])
	}

	respMsg := ScaleRespMsg{MsgType: m.SCALE_PASSTH_DATA, MsgBody: passthStr, ScaleId: scaleId}

	return respMsg, index + 2
}

var okBytes = []byte("\r\nOK\r\n")

func containsOK(response []byte) bool {
	return bytes.Contains(response, okBytes)
}

type CWLAPResponse struct {
	NetworkType int
	SSID        string
	RSSI        int
	BSSID       string
	Channel     int
	Offset      int
	Security    int
}

func convertResponseToInfo(seqNo int, response CWLAPResponse) APInfo {
	var info APInfo

	info.SeqNo = seqNo
	info.Ssid = response.SSID
	info.Rssi = getRssiLevel(response.RSSI)
	info.Mac = response.BSSID
	info.encryptType = getEncryptType(response.Security)

	return info
}

func getRssiLevel(rssi int) int {
	signalQuality := dBmtoPercentage(rssi)
	switch {
	case signalQuality < 25:
		return 1
	case signalQuality >= 25 && signalQuality < 50:
		return 2
	case signalQuality >= 50 && signalQuality < 75:
		return 3
	case signalQuality >= 75 && signalQuality <= 100:
		return 4
	default:
		return 1
	}
}

func dBmtoPercentage(rssiDbm int) int { // -50 - -100dbm
	var quality int
	if rssiDbm <= RSSI_MIN {
		quality = 0
	} else if rssiDbm >= RSSI_MAX {
		quality = 100
	} else {
		quality = 2 * (rssiDbm + 100)
	}

	return quality
} // dBmtoPercentage

func getEncryptType(security int) string {
	switch security {
	case 1:
		return "NONE"
	case 2:
		return "WEP"
	case 3:
		return "WPA"
	case 4:
		return "WPA2-PSK"
	default:
		return ""
	}
}

func parseCWLAPResponse(data []byte) []CWLAPResponse {
	if !containsOK(data) {
		return nil
	}
	response := bytes.NewBuffer(data).String()
	lines := strings.Split(response, "\n")
	regex := regexp.MustCompile(`\+CWLAP:\(([^)]+)\)`)

	var results []CWLAPResponse

	for _, line := range lines {
		match := regex.FindStringSubmatch(line)
		if len(match) > 0 {
			fields := strings.Split(match[1], ",")
			if len(fields) == 7 {
				networkType, _ := strconv.Atoi(fields[0])
				ssid := strings.Trim(fields[1], "\"")
				rssi, _ := strconv.Atoi(fields[2])
				bssid := strings.Trim(fields[3], "\"")
				channel, _ := strconv.Atoi(fields[4])
				offset, _ := strconv.Atoi(fields[5])
				security, _ := strconv.Atoi(fields[6])

				result := CWLAPResponse{
					NetworkType: networkType,
					SSID:        ssid,
					RSSI:        rssi,
					BSSID:       bssid,
					Channel:     channel,
					Offset:      offset,
					Security:    security,
				}

				results = append(results, result)
			}
		}
	}

	return results
}

func convertResponsesToInfos(responses []CWLAPResponse) []APInfo {
	infos := make([]APInfo, len(responses))

	for i, response := range responses {
		infos[i] = convertResponseToInfo(i, response)
	}

	return infos
}

func handleWifiPassthResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	switch GExpectWifiResp {
	case m.GET_AP_LIST_RESP:
		return handleGetApListResp(scaleId, data)
	case m.CONNECT_AP_RESP:
		return handleConnectApResp(scaleId, data)
	case m.CONNECT_AP_ONE_KEY_RESP:
		return handleConnectApOneKeyResp(scaleId, data)
	case m.SET_WIFI_DYNAMIC_IP_RESP:
		return handleSetWifiDynamicIpResp(scaleId, data)
	case m.SEND_DATA_TO_WIFI_RESP:
		return handleSendDataToWifiResp(scaleId, data)
	case m.GET_WIFI_AP_INFO_RESP:
		return handleGetApInfoResp(scaleId, data)
	case m.GET_IP_INFO_RESP:
		return handleGetIpInfoResp(scaleId, data)
	case m.GET_IP_MODE_RESP:
		return handleGetIpModeResp(scaleId, data)
	case m.SET_WIFI_STATIC_IP_RESP:
		return handleSetWifiStaticIpResp(scaleId, data)
	case m.CHANGE_WIFI_MODE_RESP:
		return handleChangeWifiModeResp(scaleId, data)

	default:
		return ScaleRespMsg{}, 0
	}
}

func handleBTPassthResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	switch GExpectBTResp {
	case m.MODIFY_BT_NAME_RESP:
		return handleModifyBtNameResp(scaleId, data)
	case m.BT_PASSTH_DATA_RESP:
		return handleSendDataToBTResp(scaleId, data)
	default:
		return ScaleRespMsg{}, 0
	}
}

func handleGetApListResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	if !containsOK(data) {
		return ScaleRespMsg{}, 0
	}

	apList := parseCWLAPResponse(data)
	if apList == nil {
		return ScaleRespMsg{}, 0
	}

	// compose response
	apInfoList := convertResponsesToInfos(apList)

	jsonData, err := json.MarshalToString(apInfoList)
	if err != nil {
		fmt.Println("Error:", err)
		return ScaleRespMsg{}, 0
	}
	return ScaleRespMsg{ScaleId: scaleId, MsgType: m.GET_AP_LIST_RESP, MsgBody: jsonData}, len(data)
}

func handleConnectApResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	if strings.Contains(string(data), CONNECT_AP_OK_RESP) { // success
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.CONNECT_AP_RESP, MsgBody: "ok"}, len(data)
	} else if strings.Contains(string(data), CONNECT_AP_FAIL_RESP) { // fail
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.CONNECT_AP_RESP, MsgBody: "fail"}, len(data)
	} else { // unkown
		return ScaleRespMsg{}, len(data)
	}
}

func handleConnectApOneKeyResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	if strings.Contains(string(data), CONNECT_AP_OK_RESP) { // success
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.CONNECT_AP_ONE_KEY_RESP, MsgBody: "ok"}, len(data)
	} else if strings.Contains(string(data), CONNECT_AP_FAIL_RESP) { // fail
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.CONNECT_AP_ONE_KEY_RESP, MsgBody: "fail"}, len(data)
	} else { // unkown
		return ScaleRespMsg{}, len(data)
	}
}

func handleChangeWifiModeResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	if strings.Contains(string(data), CHANG_WIFI_MODE_OK_RESP) { // success
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.CHANGE_WIFI_MODE_RESP, MsgBody: "ok"}, len(data)
	} else if strings.Contains(string(data), CONNECT_AP_FAIL_RESP) { // fail
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.CHANGE_WIFI_MODE_RESP, MsgBody: "fail"}, len(data)
	} else { // unkown
		return ScaleRespMsg{}, len(data)
	}
}

func handleRescanApListResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	// TODO: Implement function
	return ScaleRespMsg{}, 0
}

func handleSetWifiDynamicIpResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	if strings.Contains(string(data), SET_WIFI_DYNAMIC_IP_OK_RESP) { // success
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.SET_WIFI_DYNAMIC_IP_RESP, MsgBody: "ok"}, len(data)
	} else { // unkown
		return ScaleRespMsg{}, 0
	}
}

func handleSetWifiStaticIpResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	if strings.Contains(string(data), SET_WIFI_STATIC_IP_OK_RESP) { // success
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.SET_WIFI_STATIC_IP_RESP, MsgBody: "ok"}, len(data)
	} else {
		return ScaleRespMsg{}, 0
	}
}

type IPInfo struct {
	IP      string
	Gateway string
	Netmask string
}

type WifiAPInfo struct {
	Ssid    string
	Bssid   string
	Channel string
	Rssi    int
}

func extractWifiAPInfo(response string) (WifiAPInfo, error) {
	var info WifiAPInfo

	// 	+CWJAP_DEF:<ssid>, <bssid>, <channel>, <rssi>
	// OK
	// split response string into multiple lines
	lines := strings.Split(response, "\n")

	// iterates on each lines to extract ssid, bssid, channel, rssi
	for _, line := range lines {
		line = strings.Trim(line, "\t")
		line = strings.Replace(line, `\"`, "", -1)
		line = strings.Replace(line, `"`, "", -1)
		if strings.HasPrefix(line, "+CWJAP_DEF:") {
			data := line[len("+CWJAP_DEF:"):]
			dataSplit := strings.Split(data, ",")
			info.Ssid = dataSplit[0]
			info.Bssid = dataSplit[1]
			info.Channel = dataSplit[2]
			level, _ := strconv.ParseInt(dataSplit[3], 10, 64)
			info.Rssi = getRssiLevel(int(level))
		}
	}

	return info, nil
}

func extractIPInfo(response string) (IPInfo, error) {
	var info IPInfo

	// 根据字符串中的换行符分割字符串
	lines := strings.Split(response, "\r\n")

	// 遍历每一行字符串，提取 IP、网关和子网掩码的值
	for _, line := range lines {
		if strings.HasPrefix(line, "+CIPSTA_CUR:ip:") {
			info.IP = strings.Trim(line[len("+CIPSTA_CUR:ip:\""):], "\"")
		} else if strings.HasPrefix(line, "+CIPSTA_CUR:gateway:") {
			info.Gateway = strings.Trim(line[len("+CIPSTA_CUR:gateway:\""):], "\"")
		} else if strings.HasPrefix(line, "+CIPSTA_CUR:netmask:") {
			info.Netmask = strings.Trim(line[len("+CIPSTA_CUR:netmask:\""):], "\"")
		}
	}

	// 检查是否成功提取了所有值
	if info.IP == "" || info.Gateway == "" || info.Netmask == "" {
		return info, fmt.Errorf("Failed to extract IP information")
	}

	return info, nil
}

func extractIPMode(response string) (bool, error) {
	var mode bool
	var err error = nil

	// 根据字符串中的换行符分割字符串
	lines := strings.Split(response, "\r\n")

	// 遍历每一行字符串，提取 IP mode
	for _, line := range lines {
		if strings.HasPrefix(line, "+CWDHCP_CUR:") {
			modeNo := line[len("+CWDHCP_CUR:"):]
			if modeNo == "2" || modeNo == "3" {
				mode = true
			} else if modeNo == "0" || modeNo == "1" {
				mode = false
			} else {
				mode = false
				err = fmt.Errorf("invalid response")
			}
			break
		}
	}

	return mode, err
}

func handleGetApInfoResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	if strings.Contains(string(data), GET_AP_INFO_OK_RESP) { // success
		apInfo, err := extractWifiAPInfo(string(data))
		if err != nil {
			return ScaleRespMsg{}, len(data)
		}
		apInfoStr, _ := json.MarshalToString(apInfo)
		return ScaleRespMsg{m.GET_WIFI_AP_INFO_RESP, apInfoStr, scaleId}, len(data)
	} else if strings.Contains(string(data), "Error") { // fail
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.GET_WIFI_AP_INFO_RESP, MsgBody: "fail"}, len(data)
	} else { // unkown
		return ScaleRespMsg{}, 0
	}
}

func handleGetIpInfoResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	if strings.Contains(string(data), GET_IP_INFO_OK_RESP) { // success
		ipInfo, err := extractIPInfo(string(data))
		if err != nil {
			return ScaleRespMsg{}, len(data)
		}
		ipInfoStr, _ := json.MarshalToString(ipInfo)
		return ScaleRespMsg{m.GET_IP_INFO_RESP, ipInfoStr, scaleId}, len(data)
	} else if strings.Contains(string(data), "Error") { // fail
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.GET_IP_INFO_RESP, MsgBody: "fail"}, len(data)
	} else { // unkown
		return ScaleRespMsg{}, 0
	}
}

func handleGetIpModeResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	if strings.Contains(string(data), GET_IP_MODE_OK_RESP) { // success
		isDhcpEnabled, err := extractIPMode(string(data))
		if err != nil {
			return ScaleRespMsg{}, len(data)
		}
		var dhcpStr string
		if isDhcpEnabled {
			dhcpStr = "dhcp"
		} else {
			dhcpStr = "static"
		}

		return ScaleRespMsg{m.GET_IP_MODE_RESP, dhcpStr, scaleId}, len(data)
	} else if strings.Contains(string(data), "Error") { // fail
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.GET_IP_MODE_RESP, MsgBody: "fail"}, len(data)
	} else { // unkown
		return ScaleRespMsg{}, 0
	}
}

func handleModifyBtNameResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	if bytes.Contains(data, []byte(MODIFY_BT_OK_RESP)) {
		return ScaleRespMsg{m.MODIFY_BT_NAME_RESP, "ok", scaleId}, len(data)
	} else {
		return ScaleRespMsg{m.MODIFY_BT_NAME_RESP, "fail", scaleId}, len(data)
	}
}

func handleSendDataToBTResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	return ScaleRespMsg{m.SEND_DATA_TO_BT_RESP, string(data), scaleId}, len(data)
}

func handleSendDataToWifiResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	return ScaleRespMsg{m.SEND_DATA_TO_WIFI_RESP, string(data), scaleId}, len(data)
}
