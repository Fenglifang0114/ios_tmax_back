package svc

import (
	"fmt"
	"strconv"

	"tmaxsrv/log"
	"tmaxsrv/utils"
)

type reqProcFun func(scale *Scale, req SRequest) error

var handlers map[SReqType]reqProcFun

func init() {
	handlers = map[SReqType]reqProcFun{
		SREQ_GET_WEIGHT:          procGetWeight,
		SREQ_ZERO:                procZero,
		SREQ_TARE:                procTare,
		SREQ_REG_WEIGHT_DATA:     procRegWeight,
		SREQ_UNREG_WEIGHT_DATA:   procUnRegWeight,
		SREQ_GET_RECS:            procGetRecs,
		SREQ_ADD_REC:             procAddRec,
		SREQ_DEL_REC:             procDelRec,
		SREQ_DOWN_PRN_FMT:        procDownPrnFmt,
		SREQ_GET_AP_LIST:         procGetApList,
		SREQ_RESCAN_AP_LIST:      procRescanAp,
		SREQ_CONNECT_AP:          procConnectAp,
		SREQ_SET_WIFI_DYNAMIC_IP: procSetWifiDynamicIp,
		SREQ_SET_WIFI_STATIC_IP:  procSetWifiStaticIp,
		SREQ_GET_IP_INFO:         procGetIpInfo,
		SREQ_MODIFY_BT_NAME:      procModifyBTName,
		SREQ_SEND_DATA_TO_BT:     procSendDataToBT,
		SREQ_SEND_DATA_TO_WIFI:   procSendDataToWifi,
	}
}

func procToScaleReq(scale *Scale, req SRequest) error {
	return procReq(scale, req)
}

func procReq(scale *Scale, req SRequest) error {
	handler, ok := handlers[req.Req]
	if !ok {
		return fmt.Errorf("unknown req: %v", req)
	}

	return handler(scale, req)
}

func procGetRecs(scale *Scale, req SRequest) error {
	recs, _ := scale.GetRecs()
	recsStr, _ := json.MarshalToString(recs)
	resp := &ScaleRespMsg{MsgType: GET_RECS_RESP, MsgBody: recsStr, ScaleId: scale.Id}
	result, _ := json.Marshal(resp)
	scale.client.sendCh <- result
	return nil
}

func procAddRec(scale *Scale, req SRequest) error {
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
	return nil
}

func procDelRec(scale *Scale, req SRequest) error {
	var id uint64
	var err error
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
	return nil
}

func procDownPrnFmt(scale *Scale, req SRequest) error {
	var resp *ScaleRespMsg
	if err := ReqDownPrnFmt(scale, req.ReqData); err != nil {
		resp = &ScaleRespMsg{MsgType: DOWN_PRN_FMT_RESP, MsgBody: err.Error(), ScaleId: scale.Id}
	} else {
		resp = &ScaleRespMsg{MsgType: DOWN_PRN_FMT_RESP, MsgBody: "ok", ScaleId: scale.Id}
	}
	result, _ := json.Marshal(resp)
	scale.client.sendCh <- result
	return nil
}

func procGetApList(scale *Scale, req SRequest) error {
	if err := ReqGetApList(scale); err != nil {
		msg := ScaleRespMsg{ScaleId: scale.Id, MsgType: GET_AP_LIST_RESP, MsgBody: err.Error()}
		msgStr, _ := json.MarshalToString(msg)
		scale.client.sendCh <- []byte(msgStr)
		return err
	}
	return nil
}

func procRescanAp(scale *Scale, req SRequest) error {
	msg := ScaleRespMsg{ScaleId: scale.Id, MsgType: RESCAN_AP_LIST_RESP, MsgBody: "ok"}
	msgStr, _ := json.MarshalToString(msg)
	scale.client.sendCh <- []byte(msgStr)

	return nil
}

func procConnectAp(scale *Scale, req SRequest) error {
	data := utils.JsonToMap(req.ReqData)
	if err := ReqConnectAp(scale, data["ssid"].(string), data["bssid"].(string), data["password"].(string)); err != nil {
		err = fmt.Errorf("connect AP error")
		msg := ScaleRespMsg{ScaleId: scale.Id, MsgType: CONNECT_AP_RESP, MsgBody: err.Error()}
		msgStr, _ := json.MarshalToString(msg)
		scale.client.sendCh <- []byte(msgStr)
		return nil
	}
	// msg := ScaleRespMsg{ScaleId: scale.Id, MsgType: CONNECT_AP_RESP, MsgBody: "ok"} // TODO: restrieve AP list
	// msgStr, _ := json.MarshalToString(msg)
	// scale.client.sendCh <- []byte(msgStr)
	return nil
}

func procSetWifiDynamicIp(scale *Scale, req SRequest) error {
	return SetWifiDynamicIp(scale)
}

func procSetWifiStaticIp(scale *Scale, req SRequest) error {
	data := utils.JsonToMap(req.ReqData)
	if err := ReqSetWifiStaticIp(scale, data["ip"].(string), data["gateway"].(string), data["netmask"].(string)); err != nil {
		err = fmt.Errorf("set wifi static ip error")
		msg := ScaleRespMsg{ScaleId: scale.Id, MsgType: SET_WIFI_STATIC_IP_RESP, MsgBody: err.Error()}
		msgStr, _ := json.MarshalToString(msg)
		scale.client.sendCh <- []byte(msgStr)
		return nil
	}
	// msg := ScaleRespMsg{ScaleId: scale.Id, MsgType: SET_WIFI_STATIC_IP_RESP, MsgBody: "ok"}
	// msgStr, _ := json.MarshalToString(msg)
	// scale.client.sendCh <- []byte(msgStr)
	return nil
}

func procGetIpInfo(scale *Scale, req SRequest) error {
	if err := ReqGetIpInfo(scale); err != nil {
		err = fmt.Errorf("connect AP error")
		msg := ScaleRespMsg{ScaleId: scale.Id, MsgType: GET_IP_INFO_RESP, MsgBody: err.Error()}
		msgStr, _ := json.MarshalToString(msg)
		scale.client.sendCh <- []byte(msgStr)
		return nil
	}
	// msg := ScaleRespMsg{ScaleId: scale.Id, MsgType: GET_IP_INFO_RESP, MsgBody: "ok"} // TODO: restrieve AP list
	// msgStr, _ := json.MarshalToString(msg)
	// scale.client.sendCh <- []byte(msgStr)
	return nil
	//	msg := ScaleRespMsg{ScaleId: scale.Id, MsgType: GET_IP_INFO_RESP, MsgBody: ncnfStr}
	//	msgStr, _ := json.MarshalToString(msg)
	//	scale.client.sendCh <- []byte(msgStr)
	return nil
}

func procModifyBTName(scale *Scale, req SRequest) error {
	if err := ReqModifyBTName(scale, req.ReqData); err != nil {
		err = fmt.Errorf("modify BT name error")
		msg := ScaleRespMsg{ScaleId: scale.Id, MsgType: MODIFY_BT_NAME_RESP, MsgBody: err.Error()}
		msgStr, _ := json.MarshalToString(msg)
		scale.client.sendCh <- []byte(msgStr)
		return nil
	}
	msg := ScaleRespMsg{ScaleId: scale.Id, MsgType: MODIFY_BT_NAME_RESP, MsgBody: "ok"}
	msgStr, _ := json.MarshalToString(msg)
	scale.client.sendCh <- []byte(msgStr)
	return nil
}

func procSendDataToBT(scale *Scale, req SRequest) error {
	_ = ReqSendDataToBT(scale, req.ReqData)
	return nil
}

func procSendDataToWifi(scale *Scale, req SRequest) error {
	_ = ReqSendDataToWifi(scale, req.ReqData)
	return nil
}

func procRegWeight(scale *Scale, req SRequest) error {
	if !scale.RegWeightData() {
		return fmt.Errorf("RegWeight failed")
	}

	return nil
}

func procUnRegWeight(scale *Scale, req SRequest) error {
	if !scale.UnRegWeightData() {
		return fmt.Errorf("UnRegWeight failed")
	}

	return nil
}

func procTare(scale *Scale, req SRequest) error {
	if ok := scale.PerfTare(); !ok {
		return fmt.Errorf("perform tare error")
	}

	return nil
}

func procGetWeight(scale *Scale, req SRequest) error {
	// if scale.isBusy {
	// 	srvMgr.recvScaleMsg <- &ScaleRespMsg{ScaleId: scaleId, MsgType: WEIGHT_DATA_RESP, MsgBody: fmt.Errorf("scale is busy")}
	// 	return nil
	// }
	ok := scale.ReadWeight()
	if !ok {
		err := fmt.Errorf("get weight error")
		msg := ScaleRespMsg{ScaleId: scale.Id, MsgType: WEIGHT_DATA_RESP, MsgBody: err.Error()}
		msgStr, _ := json.MarshalToString(msg)
		scale.client.sendCh <- []byte(msgStr)
		// srvMgr.recvScaleMsg <- &ScaleRespMsg{ScaleId: scaleId, MsgType: WEIGHT_DATA_RESP, MsgBody: err.Error()}
		return nil
	}
	return nil
}

func procZero(scale *Scale, req SRequest) error {
	if ok := scale.PerfZero(); !ok {
		return fmt.Errorf("perform zero error")
	}
	return nil
}
