package svc

import (
	"fmt"
	"strconv"

	m "tmaxsrv/comm"
	"tmaxsrv/log"
	l "tmaxsrv/log"
	utils "tmaxsrv/util"
)

type (
	reqProcFun func(scale *Scale, req SRequest) (*ScaleRespMsg, error)
	IpMode     int
)

const (
	IPMODE_NONE IpMode = iota
	IPMODE_STATIC
	IPMODE_DYNAMIC
)

var handlers map[SReqType]reqProcFun

func procToScaleReq(s *Scale, req SRequest) {
	var resp *ScaleRespMsg
	var err error

	handler, ok := handlers[req.Req]
	if !ok {
		resp = &ScaleRespMsg{MsgType: GetResVsResp(req.Req), MsgBody: fmt.Sprintf("unknown req: %s", req.Req), ScaleId: s.Id}
	} else {
		resp, err = handler(s, req)
		if err != nil {
			resp = &ScaleRespMsg{MsgType: GetResVsResp(req.Req), MsgBody: fmt.Sprintf("error: %s", err.Error()), ScaleId: s.Id}
		}
	}
	// send msg to web socket client
	result, _ := json.Marshal(resp)
	s.client.sendCh <- result
}

var conversionMap map[SReqType]m.RespMsgType

func procGetRecs(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	recs, _ := scale.GetRecs()
	recsStr, _ := json.MarshalToString(recs)
	resp := &ScaleRespMsg{MsgType: m.GET_RECS_RESP, MsgBody: recsStr, ScaleId: scale.Id}
	result, _ := json.Marshal(resp)
	scale.client.sendCh <- result
	return nil, nil
}

func procAddRec(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	var rec ScaleRec
	if err := json.Unmarshal([]byte(req.ReqData), &rec); err != nil {
		log.Log.Errorf("Unmarshal ScaleRec error: %v", err)
		resp := &ScaleRespMsg{MsgType: m.GET_RECS_RESP, MsgBody: err.Error(), ScaleId: scale.Id}
		result, _ := json.Marshal(resp)
		scale.client.sendCh <- result
	} else {
		scale.AddRec(rec)
		resp := &ScaleRespMsg{MsgType: m.ADD_REC_RESP, MsgBody: "", ScaleId: scale.Id}
		result, _ := json.Marshal(resp)
		scale.client.sendCh <- result
	}
	return nil, nil
}

func procDelRec(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	var id uint64
	var err error
	if id, err = strconv.ParseUint(req.ReqData, 10, 64); err != nil {
		log.Log.Errorf(err.Error())
		resp := &ScaleRespMsg{MsgType: m.DEL_REC_RESP, MsgBody: err.Error(), ScaleId: scale.Id}
		result, _ := json.Marshal(resp)
		scale.client.sendCh <- result
	} else {
		_ = scale.DelRec(uint(id))
		resp := &ScaleRespMsg{MsgType: m.DEL_REC_RESP, MsgBody: "", ScaleId: scale.Id}
		result, _ := json.Marshal(resp)
		scale.client.sendCh <- result
	}
	return nil, nil
}

func procDownPrnFmt(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return ReqDownPrnFmt(scale, req)
}

func procGetApList(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return ReqGetApList(scale)
}

func procRescanAp(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return &ScaleRespMsg{ScaleId: scale.Id, MsgType: m.RESCAN_AP_LIST_RESP, MsgBody: "ok"}, nil // FIXME:
}

func procConnectAp(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	data := utils.JsonToMap(req.ReqData)
	return ReqConnectAp(scale, data["ssid"].(string), data["bssid"].(string), data["password"].(string))
}

func procSetWifiDynamicIp(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return ReqSetWifiDynamicIp(scale)
}

func procSetWifiStaticIp(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	data := utils.JsonToMap(req.ReqData)
	return ReqSetWifiStaticIp(scale, data["ip"].(string), data["gateway"].(string), data["netmask"].(string))
}

func procGetIpInfo(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return ReqGetIpInfo(scale)
}

func procModifyBTName(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return ReqModifyBTName(scale, req.ReqData)
}

func procSendDataToBT(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return ReqSendDataToBT(scale, req.ReqData)
}

func procGetIpMode(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return ReqGetIpMode(scale)
}

func procGetWifiInfo(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return ReqGetWifiApInfo(scale)
}

func procSendDataToWifi(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return ReqSendDataToWifi(scale, req.ReqData)
}

func procRegWeight(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	l.Log.Debugf("process register weight data")
	return scale.RegWeightData()
}

func procUnRegWeight(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return scale.UnRegWeightData()
}

func procTare(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return scale.PerfTare()
}

func procGetWeight(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return scale.ReadWeight()
}

func procZero(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return scale.PerfZero()
}

func GetResVsResp(reqType SReqType) m.RespMsgType {
	return conversionMap[reqType]
}

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
		SREQ_GET_IP_MODE:         procGetIpMode,
		SREQ_GET_WIFI_INFO:       procGetWifiInfo,
	}

	conversionMap = map[SReqType]m.RespMsgType{
		SREQ_ZERO:                m.ZERO_CMD_RESP,
		SREQ_TARE:                m.TARE_CMD_RESP,
		SREQ_GET_WEIGHT:          m.WEIGHT_DATA_RESP,
		SREQ_SEND_WT_CONT:        m.WEIGHT_DATA_RESP,
		SREQ_STOP_SEND_WT:        m.WEIGHT_DATA_RESP,
		SREQ_REG_WEIGHT_DATA:     m.REG_WEIGHT_RESP,
		SREQ_UNREG_WEIGHT_DATA:   m.UNREG_WEIGHT_RESP,
		SREQ_GET_RECS:            m.GET_RECS_RESP,
		SREQ_ADD_REC:             m.ADD_REC_RESP,
		SREQ_DEL_REC:             m.DEL_REC_RESP,
		SREQ_DOWN_PRN_FMT:        m.DOWN_PRN_FMT_RESP,
		SREQ_GET_AP_LIST:         m.GET_AP_LIST_RESP,
		SREQ_RESCAN_AP_LIST:      m.RESCAN_AP_LIST_RESP,
		SREQ_CONNECT_AP:          m.CONNECT_AP_RESP,
		SREQ_SET_WIFI_DYNAMIC_IP: m.SET_WIFI_DYNAMIC_IP_RESP,
		SREQ_SET_WIFI_STATIC_IP:  m.SET_WIFI_STATIC_IP_RESP,
		SREQ_GET_IP_INFO:         m.GET_IP_INFO_RESP,
		SREQ_MODIFY_BT_NAME:      m.MODIFY_BT_NAME_RESP,
		SREQ_SEND_DATA_TO_BT:     m.SEND_DATA_TO_BT_RESP,
		SREQ_SEND_DATA_TO_WIFI:   m.SEND_DATA_TO_WIFI_RESP,
		SREQ_GET_IP_MODE:         m.GET_IP_INFO_RESP,
		SREQ_GET_WIFI_INFO:       m.GET_IP_INFO_RESP,
	}
}
