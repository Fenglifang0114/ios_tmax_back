package svc

import (
	"fmt"
	"regexp"
	"strings"
	// m "tmaxsrv/comm"
)

// func convToMspType(inChar []byte) m.RespMsgType {
// 	switch inChar[3] {
// 	case 0x05:
// 		return m.REG_WEIGHT_RESP // FIXME:
// 	case 0xF2:

// 		switch inChar[4] {
// 		case 0x01:
// 			return m.BT_PASSTH_DATA_RESP
// 		case 0x02:
// 			return m.WIFI_PASSTH_DATA_RESP
// 		case 0x03:
// 			return m.PRT_PASSTH_DATA_RESP
// 		default:
// 			return m.REG_WEIGHT_RESP // FIXME:
// 		}

// 	}
// 	return m.REG_WEIGHT_RESP // FIXME:
// }

// return WeightMsg package and return the len of parsed in data
func retreiveWeightC51(data []byte) (pack WeightMsg, err error) {
	weightMsg := WeightMsg{}
	dataStr := string(data)
	re := regexp.MustCompile(`(\w+),(\w+),\s*([\-\d.]+)(\w+|%)|(--UL--|--OL--)`)

	matches := re.FindAllStringSubmatch(dataStr, -1)
	if matches == nil || len(matches[0]) == 0 {
		return weightMsg, fmt.Errorf("parse error on %v", string(data))
	}
	if matches[0][len(matches[0])-1] != "" { // --OL--, --UL--
		weightMsg.IsStable = false
		weightMsg.IsNet = false
		weightMsg.WeightVal = matches[0][len(matches[0])-1]
		weightMsg.WeightUnit = ""
	} else {
		// fields will be "ST,NT, 5.123kg" or "ST,NT,-0.123kg", or "-- UL --", "-- OL --"
		weightMsg.IsStable = strings.Contains(matches[0][1], "ST")
		weightMsg.IsNet = strings.Contains(matches[0][2], "NT")
		weightMsg.WeightVal = matches[0][3]
		weightMsg.WeightUnit = matches[0][4]
	}
	return weightMsg, nil
}

// func extractMsgOldScale(data []byte) (*ScaleRespMsg, error) {
// 	// TODO: check if data is response or unsolicited message, our C51 MCU scale will not response any character for T/Z/W
// 	msg, err := retreiveWeightC51(data)
// 	if err != nil { // no packet ready, just return
// 		log.Log.Info(err.Error())
// 		return nil, nil
// 	}

// 	scaleMsg := ScaleRespMsg{}
// 	scaleMsg.MsgType = m.WEIGHT_DATA
// 	scaleMsg.MsgBody = msg
// 	return &scaleMsg, err
// }

// func extractMsgTmaxScale(scaleId int64, data []byte) (*ScaleRespMsg, error) {
// 	// TODO: check if data is response or unsolicited message, our C51 MCU scale will not response any character for T/Z/W
// 	msg, err := retreiveRespMsg(scaleId, data)
// 	if err != nil { // no packet ready, just return
// 		log.Log.Info(err.Error())
// 		return nil, err
// 	}

// 	return msg, err
// }

// func retreiveRespMsg(scaleId int64, data []byte) (*ScaleRespMsg, error) {
// 	// checkHead, get msgid, get msgtype, check if return code is 0x06, for success
// 	var err error
// 	respMsg := &ScaleRespMsg{MsgType: GlastWantRespMsgType, MsgBody: "", ScaleId: scaleId}
// 	// should check the response type at [3]
// 	if data[0] == 0x5a && data[1] == 0xa5 {
// 		respMsgType := convToMspType(data)
// 		fmt.Println(respMsgType)
// 		if data[4] == 0x06 {
// 			respMsg.MsgBody = "ok"
// 			err = nil
// 		} else if data[4] == 0x15 {
// 			respMsg.MsgBody = "fail"
// 			err = nil
// 		} else if data[4] == 0x7f {
// 			respMsg.MsgType = m.ERR_SERIAL_RESP
// 			respMsg.MsgBody = "serial port error"
// 			err = nil
// 		}
// 	} else {
// 		var msg WeightMsg
// 		if msg, err = retreiveWeightC51(data); err == nil {
// 			respMsg.MsgType = m.WEIGHT_DATA
// 			respMsg.MsgBody = msg
// 		}
// 	}
// 	return respMsg, err
// }
