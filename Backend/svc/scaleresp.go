package svc

import (
	"fmt"
	"regexp"
	"strings"

	"tmaxsrv/log"
)

func convToMspType(inChar []byte) RespMsgType {
	switch inChar[3] {
	case 0x05:
		return REG_WEIGHT_RESP // FIXME:
	case 0xF2:
		if inChar[4] == 0x01 {
			return BT_PASSTH_DATA_RESP
		} else if inChar[4] == 0x02 {
			return WIFI_PASSTH_DATA_RESP
		} else if inChar[4] == 0x03 {
			return PRT_PASSTH_DATA_RESP
		} else {
			return REG_WEIGHT_RESP // FIXME:
		}
	}
	return REG_WEIGHT_RESP // FIXME:
}

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
	// should check the response type at [3]
	if data[0] == 0x5a && data[1] == 0xa5 {
		respMsgType := convToMspType(data)
		fmt.Println(respMsgType)
		if data[4] == 0x06 {
			respMsg.MsgBody = "ok"
			err = nil
		} else if data[4] == 0x15 {
			respMsg.MsgBody = "fail"
			err = nil
		} else if data[4] == 0x7f {
			respMsg.MsgType = ERR_SERIAL_RESP
			respMsg.MsgBody = "serial port error"
			err = nil
		}
	} else {
		var msg WeightMsg
		if msg, err = retreiveWeightC51(data); err == nil {
			respMsg.MsgType = WEIGHT_DATA
			respMsg.MsgBody = msg
		}
	}
	return respMsg, err
}
