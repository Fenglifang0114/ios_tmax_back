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

// NHB新增的  20250905
// 1. 标准格式: ST,NT- 123.45kg 或 ST,NT 123.45kg
// 2. 多逗号格式: ST,NT,-    123g  或 ST,NT,    123g
// 3. 不稳定数据: ------或者--OL--后者--UL--
// ST,GS    0.0(0) ct格式
// ST,GS    0.0(0) g 格式
// ST,NT- 492.5(5) ct格式
// ST,NT-   98.5g  格式
// ST,GS     0.0g  格式
func retreiveWeightNewC51(data []byte) (pack WeightMsg, err error) {
	weightMsg := WeightMsg{}
	dataStr := string(data)

	//先去掉左括号和右括号
	dataStr = strings.ReplaceAll(dataStr, "(", "")
	dataStr = strings.ReplaceAll(dataStr, ")", "")

	re := regexp.MustCompile(`^([A-Z]{2}),([A-Z]{2})(,?)\s*([+-]?)\s+([0-9]+\.[0-9]+(?:\.[0-9]+\.[0-9]+)?|[0-9]+)\s*([a-zA-Z%]+)\s*$|^(--(?:OL|UL)--|-{6,})\s*$`)
	matches := re.FindAllStringSubmatch(dataStr, -1)
	if matches == nil || len(matches[0]) < 7 {
		return weightMsg, fmt.Errorf("parse error on %v", string(data))
	}

	// 检查是否匹配到全"------"格式
	if matches[0][7] != "" { // 全"------"情况
		weightMsg.IsStable = false
		weightMsg.IsNet = false
		weightMsg.IsZero = false
		weightMsg.WeightVal = matches[0][7]
		weightMsg.WeightUnit = ""
	} else {
		status := matches[0][1] // ST, US 等
		mode := matches[0][2]   // GS, NT 等
		sign := matches[0][4]   // 正负号
		value := matches[0][5]  // 重量值
		unit := matches[0][6]   // 单位
		isZero := (value == "0" || value == "0.0" || value == "0.00" || value == "0.000" || value == "0.0000")

		// 处理正负号
		if sign == "-" {
			value = "-" + value
		}

		weightMsg.IsStable = (status == "ST")
		weightMsg.IsNet = (mode == "NT")
		weightMsg.WeightVal = value
		weightMsg.WeightUnit = unit
		weightMsg.IsZero = isZero

	}

	return weightMsg, nil
}
