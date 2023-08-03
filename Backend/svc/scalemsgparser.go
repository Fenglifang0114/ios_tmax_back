package svc

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gitteamer/log"
)

const (
	MODIFY_BT_OK_RESP string = "TTM:OK\r\n"
)
const (
	RSSI_MAX = -50  // maximum strength of signal in dBm
	RSSI_MIN = -100 // minimum strength of signal in dBm
)

var responseHandlerMap map[RespMsgType]func(int64, []byte) (ScaleRespMsg, int)

func init() {
	cmdsRespMap = CmdMap{
		0xe107: WEIGHT_DATA,
		0xe103: ZERO_CMD_RESP,
		0xe105: TARE_CMD_RESP,
		0xe101: WEIGHT_DATA_RESP,
		0xfff3: REG_WEIGHT_RESP,
		0xe108: UNREG_WEIGHT_RESP,
		0xfff6: GET_RECS_RESP,
		0xfff7: ADD_REC_RESP,
		0xfff8: DEL_REC_RESP,
		0x05f1: EN_FAC_MODE_RESP,
		0x05f2: DIS_FAC_MODE_RESP,
		0x05f3: EN_PASSTH_MODE_RESP,
		0x05f4: DIS_PASSTH_MODE_RESP,
		0xfff9: ERASE_FLASH_RESP,
		0xff10: WRITE_DATA_FLASH_RESP,
		0xff11: DOWN_PRN_FMT_RESP,
		0xff12: ERR_SERIAL_RESP,
		0xff13: GET_AP_LIST_RESP,
		0xff14: RESCAN_AP_LIST_RESP,
		0xff15: SET_WIFI_DYNAMIC_IP_RESP,
		0xff16: SET_WIFI_STATIC_IP_RESP,
		0xff17: GET_IP_INFO_RESP,
		0xf201: MODIFY_BT_NAME_RESP,
		0xff19: NO_RESP,
		0xff20: BT_PASSTH_DATA_RESP,
		0xf202: WIFI_PASSTH_DATA_RESP,
		0xff22: PRT_PASSTH_DATA_RESP,
		0xff23: UNKNOWN_DATA,
	}

	responseHandlerMap = map[RespMsgType]func(int64, []byte) (ScaleRespMsg, int){
		WEIGHT_DATA:              handleWeightDataMsg,
		ZERO_CMD_RESP:            handleZeroCmdResp,
		TARE_CMD_RESP:            handleTareCmdResp,
		WEIGHT_DATA_RESP:         handleWeightDataResp,
		REG_WEIGHT_RESP:          handleRegWeightResp,
		UNREG_WEIGHT_RESP:        handleUnregWeightResp,
		GET_RECS_RESP:            handleGetRecsResp,
		ADD_REC_RESP:             handleAddRecResp,
		DEL_REC_RESP:             handleDelRecResp,
		EN_FAC_MODE_RESP:         handleEnFacModeResp,
		DIS_FAC_MODE_RESP:        handleDisFacModeResp,
		EN_PASSTH_MODE_RESP:      handleEnPassthModeResp,
		DIS_PASSTH_MODE_RESP:     handleDisPassthModeResp,
		ERASE_FLASH_RESP:         handleEraseFlashResp,
		WRITE_DATA_FLASH_RESP:    handleWriteDataFlashResp,
		DOWN_PRN_FMT_RESP:        handleDownPrnFmtResp,
		ERR_SERIAL_RESP:          handleErrSerialResp,
		GET_AP_LIST_RESP:         handleGetApListResp,
		RESCAN_AP_LIST_RESP:      handleRescanApListResp,
		SET_WIFI_DYNAMIC_IP_RESP: handleSetWifiDynamicIpResp,
		SET_WIFI_STATIC_IP_RESP:  handleSetWifiStaticIpResp,
		GET_IP_INFO_RESP:         handleGetIpInfoResp,
		MODIFY_BT_NAME_RESP:      handleModifyBtNameResp,
		WIFI_PASSTH_DATA_RESP:    handleWifiPassthResp,
	}

	// example usage: call the handler for the WEIGHT_DATA message
	// msg := "some message"
	// responseHandlerMap[WEIGHT_DATA](msg)
}
func extractMessage(scaleId int64, bufs *CircularBuffer, msgType RespMsgType) ScaleRespMsg {
	data := bufs.PeekAll()
	handler := responseHandlerMap[msgType]
	if handler == nil {
		log.Error("handler not found, msgType: %v", msgType)
	}
	resp, shouldRemoveLen := responseHandlerMap[msgType](scaleId, data)
	bufs.DequeueN(shouldRemoveLen)

	return resp
}

func handleWeightDataMsg(scaleId int64, data []byte) (ScaleRespMsg, int) {
	weightMsg, err := retreiveWeight(data)
	if err != nil {
		return ScaleRespMsg{}, len(data)
	}
	weightStr, err := json.MarshalToString(weightMsg)
	if err == nil {
		fmt.Printf("%v", weightStr)
	}
	respMsg := ScaleRespMsg{MsgType: WEIGHT_DATA, MsgBody: weightStr, ScaleId: scaleId}

	return respMsg, len(data)
}

func retreiveWeight(data []byte) (pack WeightMsg, err error) {
	dataStr := string(data)
	// fields will be "ST,NT, 5.123kg" or "ST,NT,-0.123kg", or "-- UL --", "-- OL --"
	fields := strings.Split(dataStr, ",")
	if len(fields) != 3 {
		if len(fields[0]) < MIN_PACK_SIZE {
			return WeightMsg{}, fmt.Errorf("no packet")
		}

		return WeightMsg{WeightVal: strings.TrimRight(dataStr, "\r\n"), WeightUnit: ""}, nil
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

func handleZeroCmdResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	var msg = ScaleRespMsg{}
	if data[0] == 0x06 {
		msg.MsgType = ZERO_CMD_RESP
		msg.ScaleId = scaleId
		msg.MsgBody = "ok"
	} else {
		msg.MsgType = ZERO_CMD_RESP
		msg.ScaleId = scaleId
		msg.MsgBody = "fail"
	}
	return msg, len(data)
}

func handleTareCmdResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	var msg = ScaleRespMsg{}
	if data[0] == 0x06 {
		msg.MsgType = TARE_CMD_RESP
		msg.ScaleId = scaleId
		msg.MsgBody = "ok"
	} else {
		msg.MsgType = TARE_CMD_RESP
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
	var msg = ScaleRespMsg{}
	if data[0] == 0x06 {
		msg.MsgType = REG_WEIGHT_RESP
		msg.ScaleId = scaleId
		msg.MsgBody = "ok"
	} else {
		msg.MsgType = REG_WEIGHT_RESP
		msg.ScaleId = scaleId
		msg.MsgBody = "fail"
	}
	return msg, len(data)
}

func handleUnregWeightResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	var msg = ScaleRespMsg{}
	if data[0] == 0x06 {
		msg.MsgType = UNREG_WEIGHT_RESP
		msg.ScaleId = scaleId
		msg.MsgBody = "ok"
	} else {
		msg.MsgType = UNREG_WEIGHT_RESP
		msg.ScaleId = scaleId
		msg.MsgBody = "fail"
	}
	return msg, len(data)
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
		return ScaleRespMsg{ScaleId: scaleId, MsgType: EN_FAC_MODE_RESP, MsgBody: "ok"}, len(data)
	} else {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: EN_FAC_MODE_RESP, MsgBody: "fail"}, len(data)
	}
}

func handleDisFacModeResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	if data[0] == 0x06 {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: DIS_FAC_MODE_RESP, MsgBody: "ok"}, len(data)
	} else {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: DIS_FAC_MODE_RESP, MsgBody: "fail"}, len(data)
	}
}

func handleEnPassthModeResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	if data[0] == 0x06 {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: EN_PASSTH_MODE_RESP, MsgBody: "ok"}, len(data)
	} else {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: EN_PASSTH_MODE_RESP, MsgBody: "fail"}, len(data)
	}
}

func handleDisPassthModeResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	if data[0] == 0x06 {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: DIS_PASSTH_MODE_RESP, MsgBody: "ok"}, len(data)
	} else {
		return ScaleRespMsg{ScaleId: scaleId, MsgType: DIS_PASSTH_MODE_RESP, MsgBody: "fail"}, len(data)
	}
}

func handleEraseFlashResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	// TODO: Implement function
	return ScaleRespMsg{}, 0
}

func handleWriteDataFlashResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	// TODO: Implement function
	return ScaleRespMsg{}, 0
}

func handleDownPrnFmtResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	// TODO: Implement function
	return ScaleRespMsg{}, 0
}

func handleErrSerialResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	// TODO: Implement function
	return ScaleRespMsg{}, 0
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
} //dBmtoPercentage

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
	case "get_ap_list":
		return handleGetApListResp(scaleId, data)
	case "connect_ap":
		return handleConnectApResp(scaleId, data)
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
	return ScaleRespMsg{ScaleId: scaleId, MsgType: GET_AP_LIST_RESP, MsgBody: jsonData}, len(data)
}

func handleConnectApResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	if strings.Contains(string(data), "WIFI CONNECTED") { // success
		return ScaleRespMsg{ScaleId: scaleId, MsgType: CONNECT_AP_RESP, MsgBody: "ok"}, len(data)
	} else if strings.Contains(string(data), "WIFI DISCONNECTED") { // fail
		return ScaleRespMsg{ScaleId: scaleId, MsgType: CONNECT_AP_RESP, MsgBody: "fail"}, len(data)
	} else { // unkown
		return ScaleRespMsg{}, 0
	}
}

func handleRescanApListResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	// TODO: Implement function
	return ScaleRespMsg{}, 0
}

func handleSetWifiDynamicIpResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	// TODO: Implement function
	return ScaleRespMsg{}, 0
}

func handleSetWifiStaticIpResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	// TODO: Implement function
	return ScaleRespMsg{}, 0
}

func handleGetIpInfoResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	// TODO: Implement function
	return ScaleRespMsg{}, 0
}

func handleModifyBtNameResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	if bytes.Contains(data, []byte(MODIFY_BT_OK_RESP)) {
		return ScaleRespMsg{MODIFY_BT_NAME_RESP, "ok", scaleId}, len(data)
	} else {
		return ScaleRespMsg{MODIFY_BT_NAME_RESP, "fail", scaleId}, len(data)
	}
}

func handleNoResponse(scaleId int64, data []byte) (ScaleRespMsg, int) {
	// TODO: Implement function
	return ScaleRespMsg{}, 0
}

func handleBtPassthData(scaleId int64, data []byte) (ScaleRespMsg, int) {
	// TODO: Implement function
	return ScaleRespMsg{}, 0
}

func handleWifiPassthData(scaleId int64, data []byte) (ScaleRespMsg, int) {
	// TODO: Implement function
	return ScaleRespMsg{}, 0
}

func handlePrtPassthData(scaleId int64, data []byte) (ScaleRespMsg, int) {
	// TODO: Implement function
	return ScaleRespMsg{}, 0
}

func handleUnknownData(scaleId int64, data []byte) (ScaleRespMsg, int) {
	// TODO: Implement function
	return ScaleRespMsg{}, 0
}
