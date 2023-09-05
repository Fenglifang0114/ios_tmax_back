package svc

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gitteamer/log"

	m "tmaxsrv/comm"
	"tmaxsrv/util"
)

const (
	MODIFY_BT_OK_RESP string = "TTM:OK\r\n"
)

const (
	CONNECT_AP_OK_RESP          string = "\r\nOK\r\n"
	SET_WIFI_DYNAMIC_IP_OK_RESP string = "\r\nOK\r\n"
	SET_WIFI_STATIC_IP_OK_RESP  string = "\r\nOK\r\n"
	GET_IP_INFO_OK_RESP         string = "\r\nOK\r\n"
	GET_IP_MODE_OK_RESP         string = "\r\nOK\r\n"
)

const (
	RSSI_MAX = -50  // maximum strength of signal in dBm
	RSSI_MIN = -100 // minimum strength of signal in dBm
)

var responseHandlerMap map[m.RespMsgType]func(int64, []byte) (ScaleRespMsg, int)

func init() {
	util.CmdsRespMap = util.CmdMap{
		0xe107: m.WEIGHT_DATA,
		0xe103: m.ZERO_CMD_RESP,
		0xe105: m.TARE_CMD_RESP,
		0xe101: m.WEIGHT_DATA_RESP,
		0xfff3: m.REG_WEIGHT_RESP,
		0xe108: m.UNREG_WEIGHT_RESP,
		0xfff6: m.GET_RECS_RESP,
		0xfff7: m.ADD_REC_RESP,
		0xfff8: m.DEL_REC_RESP,
		0x05f1: m.EN_FAC_MODE_RESP,
		0x05f2: m.DIS_FAC_MODE_RESP,
		0x05f3: m.EN_PASSTH_MODE_RESP,
		0x05f4: m.DIS_PASSTH_MODE_RESP,
		0xfff9: m.ERASE_FLASH_RESP,
		0xff10: m.WRITE_DATA_FLASH_RESP,
		0xff11: m.DOWN_PRN_FMT_RESP,
		0xff12: m.ERR_SERIAL_RESP,
		0xff13: m.GET_AP_LIST_RESP,
		0xff14: m.RESCAN_AP_LIST_RESP,
		0xff15: m.SET_WIFI_DYNAMIC_IP_RESP,
		0xff16: m.SET_WIFI_STATIC_IP_RESP,
		0xff17: m.GET_IP_INFO_RESP,
		0xff18: m.MODIFY_BT_NAME_RESP,
		0xff19: m.NO_RESP,
		0xf201: m.BT_PASSTH_DATA_RESP,
		0xf202: m.WIFI_PASSTH_DATA_RESP,
		0xff22: m.PRT_PASSTH_DATA_RESP,
		0xff23: m.UNKNOWN_DATA,
	}

	responseHandlerMap = map[m.RespMsgType]func(int64, []byte) (ScaleRespMsg, int){
		m.WEIGHT_DATA:              handleWeightDataMsg,
		m.ZERO_CMD_RESP:            handleZeroCmdResp,
		m.TARE_CMD_RESP:            handleTareCmdResp,
		m.WEIGHT_DATA_RESP:         handleWeightDataResp,
		m.REG_WEIGHT_RESP:          handleRegWeightResp,
		m.UNREG_WEIGHT_RESP:        handleUnregWeightResp,
		m.GET_RECS_RESP:            handleGetRecsResp,
		m.ADD_REC_RESP:             handleAddRecResp,
		m.DEL_REC_RESP:             handleDelRecResp,
		m.EN_FAC_MODE_RESP:         handleEnFacModeResp,
		m.DIS_FAC_MODE_RESP:        handleDisFacModeResp,
		m.EN_PASSTH_MODE_RESP:      handleEnPassthModeResp,
		m.DIS_PASSTH_MODE_RESP:     handleDisPassthModeResp,
		m.ERASE_FLASH_RESP:         handleEraseFlashResp,
		m.WRITE_DATA_FLASH_RESP:    handleWriteDataFlashResp,
		m.DOWN_PRN_FMT_RESP:        handleDownPrnFmtResp,
		m.ERR_SERIAL_RESP:          handleErrSerialResp,
		m.GET_AP_LIST_RESP:         handleGetApListResp,
		m.RESCAN_AP_LIST_RESP:      handleRescanApListResp,
		m.SET_WIFI_DYNAMIC_IP_RESP: handleSetWifiDynamicIpResp,
		m.SET_WIFI_STATIC_IP_RESP:  handleSetWifiStaticIpResp,
		m.GET_IP_INFO_RESP:         handleGetIpInfoResp,
		m.GET_IP_MODE_RESP:         handleGetIpModeResp,
		m.MODIFY_BT_NAME_RESP:      handleModifyBtNameResp,
		m.BT_PASSTH_DATA_RESP:      handleBTPassthResp,
		m.WIFI_PASSTH_DATA_RESP:    handleWifiPassthResp,
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

func handleWeightDataMsg(scaleId int64, data []byte) (ScaleRespMsg, int) {
	weightMsg, err := retreiveWeight(data)
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
	case m.SET_WIFI_DYNAMIC_IP_RESP:
		return handleSetWifiDynamicIpResp(scaleId, data)
	case m.SEND_DATA_TO_WIFI_RESP:
		return handleSendDataToWifiResp(scaleId, data)
	case m.GET_IP_INFO_RESP:
		return handleGetIpInfoResp(scaleId, data)
	case m.GET_IP_MODE_RESP:
		return handleGetIpModeResp(scaleId, data)
	case m.SET_WIFI_STATIC_IP_RESP:
		return handleSetWifiStaticIpResp(scaleId, data)
	default:
		return ScaleRespMsg{}, 0
	}
}

func handleBTPassthResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	switch GExpectBTResp {
	case m.MODIFY_BT_NAME_RESP:
		return handleModifyBtNameResp(scaleId, data)
	case m.SEND_DATA_TO_BT_RESP:
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
	} else if strings.Contains(string(data), "+CWJAP:") { // fail
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.CONNECT_AP_RESP, MsgBody: "fail"}, len(data)
	} else { // unkown
		return ScaleRespMsg{}, 0
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
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.SET_WIFI_DYNAMIC_IP_RESP, MsgBody: "fail"}, len(data)
	}
}

func handleSetWifiStaticIpResp(scaleId int64, data []byte) (ScaleRespMsg, int) {
	if strings.Contains(string(data), SET_WIFI_STATIC_IP_OK_RESP) { // success
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.SET_WIFI_STATIC_IP_RESP, MsgBody: "ok"}, len(data)
	} else { // unkown
		return ScaleRespMsg{ScaleId: scaleId, MsgType: m.SET_WIFI_STATIC_IP_RESP, MsgBody: "fail"}, len(data)
	}
}

type IPInfo struct {
	IP      string
	Gateway string
	Netmask string
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
			break;
		}
	}

	return mode, err
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
