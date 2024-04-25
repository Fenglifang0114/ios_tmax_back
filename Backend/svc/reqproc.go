package svc

import (
	"fmt"
	"strconv"
	"strings"

	m "tmaxsrv/comm"
	"tmaxsrv/log"
	l "tmaxsrv/log"
	"tmaxsrv/picker"
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
	recs, _ := scale.GetRecs(req.ReqData)
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
		resp := &ScaleRespMsg{MsgType: m.ADD_REC_RESP, MsgBody: "ok", ScaleId: scale.Id} //@FLF20231027
		result, _ := json.Marshal(resp)
		scale.client.sendCh <- result
	}
	return nil, nil
}

func procDelRec(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	var id uint64
	var scaleMode uint64
	var err error

	parts := strings.Split(req.ReqData, ",")
	if id, err = strconv.ParseUint(parts[0], 10, 64); err != nil {
		log.Log.Errorf(err.Error())
		resp := &ScaleRespMsg{MsgType: m.DEL_REC_RESP, MsgBody: err.Error(), ScaleId: scale.Id}
		result, _ := json.Marshal(resp)
		scale.client.sendCh <- result
	} else if scaleMode, err = strconv.ParseUint(parts[1], 10, 64); err != nil {
		log.Log.Errorf(err.Error())
		resp := &ScaleRespMsg{MsgType: m.DEL_REC_RESP, MsgBody: err.Error(), ScaleId: scale.Id}
		result, _ := json.Marshal(resp)
		scale.client.sendCh <- result
	} else {
		_ = scale.DelRec(uint(id), uint(scaleMode), parts[2], parts[3])
		resp := &ScaleRespMsg{MsgType: m.DEL_REC_RESP, MsgBody: "", ScaleId: scale.Id}
		result, _ := json.Marshal(resp)
		scale.client.sendCh <- result
	}

	// if id, err = strconv.ParseUint(req.ReqData, 10, 64); err != nil {
	// 	log.Log.Errorf(err.Error())
	// 	resp := &ScaleRespMsg{MsgType: m.DEL_REC_RESP, MsgBody: err.Error(), ScaleId: scale.Id}
	// 	result, _ := json.Marshal(resp)
	// 	scale.client.sendCh <- result
	// }
	//  else {
	// 	_ = scale.DelRec(uint(id))
	// 	resp := &ScaleRespMsg{MsgType: m.DEL_REC_RESP, MsgBody: "", ScaleId: scale.Id}
	// 	result, _ := json.Marshal(resp)
	// 	scale.client.sendCh <- result
	// }
	return nil, nil
}

func procDownPrnFmt(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return ReqDownPrnFmt(scale, req)
}

func procDownPlu(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return ReqDownPlu(scale, req)
}

func procDelPlu(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return ReqDelPlu(scale, req)
}

func procInsertPlu(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return ReqInsertPlu(scale, req)
}

func procGetOneEepromInfo(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return ReqGetOneEepromInfo(scale, req)
}

func procGetAllEepromInfo(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return ReqGetAllEepromInfo(scale)
}

func ProcSetOutputFmt(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return ReqSetOutputFmt(scale, req)
}

func procGetApList(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return ReqGetApList(scale)
}

func procGetWeightErr(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return ReqGetWeightErr(scale)
}

func procRescanAp(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return &ScaleRespMsg{ScaleId: scale.Id, MsgType: m.RESCAN_AP_LIST_RESP, MsgBody: "ok"}, nil
}

func procConnectAp(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	data := utils.JsonToMap(req.ReqData)
	return ReqConnectAp(scale, data["ssid"].(string), data["bssid"].(string), data["password"].(string))
}

func procConnectApOneKey(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	data := utils.JsonToMap(req.ReqData)
	return ReqConnectApOneKey(scale, data["ssid"].(string), data["bssid"].(string), data["password"].(string))
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

func procChangeWifiMode(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return ReqChangeWifiMode(scale, req)
}

func procModifyEepromInfo(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return ReqModifyEepromInfo(scale, req)
}

func procModifyVarValue(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return ReqModifyVarValue(scale, req)
}

func procSetServerIp(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return ReqSetServerIp(scale, req)
}

func procEnFactoryMode(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return ReqSetServerIp(scale, req)
	// return ReqEnFactoryMode(scale)TODO:
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

func procOpenScalePassth(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	l.Log.Debugf("process Open Scale Passth")
	msg, err := scale.OpenScalePassth()
	if err == nil && msg.MsgBody == "ok" {
		scale.isScalePassth = true
		if req.ReqData == "hex" {
			scale.IsScalePassthHex = true
		} else {
			scale.IsScalePassthHex = false
		}
		picker := picker.GetPickerFn(scale.ScaleCat + 1)
		scale.MySerial.ChangePickFunc(picker)
	}

	return msg, err
}

func procCloseScalePassth(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	l.Log.Debugf("process Close Scale Passth")
	scale.isScalePassth = false
	picker := picker.GetPickerFn(scale.ScaleCat)
	scale.MySerial.ChangePickFunc(picker)
	return scale.CloseScalePassth()
}

func procChangeScalePassthMode(s *Scale, req SRequest) (*ScaleRespMsg, error) {
	l.Log.Debugf("process Open Scale Passth")
	if req.ReqData == "hex" {
		s.IsScalePassthHex = true
	} else {
		s.IsScalePassthHex = false
	}
	return &ScaleRespMsg{MsgType: m.CHANGE_SCALE_PASSTH_MODE_RESP, MsgBody: "ok", ScaleId: s.Id}, nil
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

func procUpdateFirmware(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return scale.UpdateFirmware(req.ReqData)
}

func ProcCheckSerialPort(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return scale.CheckSerialPort()
}

func procGetBuildInfo(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return scale.GetBuildInfo()
}

func procGetScaleTime(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return scale.GetScaleTime()
}
func procSetScaleTime(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return ReqSetScaleTime(scale, req)
}

func procGetScaleInfo(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return scale.GetScaleInfo()
}

func procGetFactoryInfo(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return scale.GetFactoryInfo()
}

func procGetWeighErr(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	return ReqGetWeightErr(scale)
}

func procGetUiConf(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	l.Log.Info("Got get UI Config request")
	modeInt, _ := strconv.Atoi(req.ReqData)
	modeUint := uint(modeInt)
	// TODO:需要连上之后加sn  ScaleSn
	config, _ := mSrvMgr.modeSetting.GetModeSetting(modeUint)
	configStr, _ := json.MarshalToString(config[0])
	respMsg := &ScaleRespMsg{MsgType: m.GET_UI_CONF_RESP, MsgBody: configStr, ScaleId: scale.Id}
	return respMsg, nil
}

func procUpdateUiConf(scale *Scale, req SRequest) (*ScaleRespMsg, error) {
	l.Log.Info("Got get UI Config request")
	// TODO:需要连上之后加sn  ScaleSn
	var config ModeSetting
	var respMsg *ScaleRespMsg
	if err := json.UnmarshalFromString(req.ReqData, &config); err != nil {
		l.Log.Error(err)
		respMsg = &ScaleRespMsg{MsgType: m.UPDATE_UI_CONF_RESP, MsgBody: "failed to parse update UI Config", ScaleId: scale.Id}

	} else {
		err := mSrvMgr.modeSetting.settingPb.UpdateModeSetting(config)
		if err != nil {
			l.Log.Error(err)
		}
		respMsg = &ScaleRespMsg{MsgType: m.UPDATE_UI_CONF_RESP, MsgBody: "ok", ScaleId: scale.Id}
	}
	return respMsg, nil
}

//  处理请求

func init() {
	handlers = map[SReqType]reqProcFun{
		SREQ_GET_WEIGHT:               procGetWeight,
		SREQ_ZERO:                     procZero,
		SREQ_TARE:                     procTare,
		SREQ_REG_WEIGHT_DATA:          procRegWeight,
		SREQ_UNREG_WEIGHT_DATA:        procUnRegWeight,
		SREQ_GET_RECS:                 procGetRecs,
		SREQ_ADD_REC:                  procAddRec,
		SREQ_DEL_REC:                  procDelRec,
		SREQ_DOWN_PRN_FMT:             procDownPrnFmt,
		SREQ_GET_AP_LIST:              procGetApList,
		SREQ_RESCAN_AP_LIST:           procRescanAp,
		SREQ_CONNECT_AP:               procConnectAp,
		SREQ_CONNECT_AP_ONE_KEY:       procConnectApOneKey,
		SREQ_SET_WIFI_DYNAMIC_IP:      procSetWifiDynamicIp,
		SREQ_SET_WIFI_STATIC_IP:       procSetWifiStaticIp,
		SREQ_GET_IP_INFO:              procGetIpInfo,
		SREQ_MODIFY_BT_NAME:           procModifyBTName,
		SREQ_SEND_DATA_TO_BT:          procSendDataToBT,
		SREQ_SEND_DATA_TO_WIFI:        procSendDataToWifi,
		SREQ_GET_IP_MODE:              procGetIpMode,
		SREQ_GET_WIFI_INFO:            procGetWifiInfo,
		SREQ_UPDATE_FIRMWARE:          procUpdateFirmware,
		SREQ_CHECK_SERIAL_PORT:        ProcCheckSerialPort,
		SREQ_GET_BUILD_INFO:           procGetBuildInfo,
		SREQ_GET_SCALE_TIME:           procGetScaleTime,
		SREQ_SET_SCALE_TIME:           procSetScaleTime,
		SREQ_GET_ONE_EEPROM_INFO:      procGetOneEepromInfo,
		SREQ_GET_ALL_EEPROM_INFO:      procGetAllEepromInfo,
		SREQ_SET_OUTPUT_FMT:           ProcSetOutputFmt,
		SREQ_OPNE_SCALE_PASSTHROUGH:   procOpenScalePassth,
		SREQ_CLOSE_SCALE_PASSTHROUGH:  procCloseScalePassth,
		SREQ_CHANGE_SCALE_PASSTH_MODE: procChangeScalePassthMode,
		SREQ_GET_SCALE_INFO:           procGetScaleInfo,
		SREQ_GET_FACTORY_INFO:         procGetFactoryInfo,
		SREQ_GET_WEIGHT_ERR:           procGetWeighErr,
		SREQ_DOWN_PLU:                 procDownPlu,
		SREQ_DEL_PLU:                  procDelPlu,
		SREQ_INSERT_PLU:               procInsertPlu,
		SREQ_GET_UI_CONF:              procGetUiConf,
		SREQ_UPDATE_UI_CONF:           procUpdateUiConf,
		SREQ_CHANGE_WIFI_MODE:         procChangeWifiMode,
		SREQ_MODIFY_EEPROM_INFO:       procModifyEepromInfo,
		SREQ_MODIFY_VAR_VALUE:         procModifyVarValue,
		SREQ_SET_SERVER_IP:            procSetServerIp,
		SREQ_EN_FACTORY_MODE:          procEnFactoryMode,
	}

	conversionMap = map[SReqType]m.RespMsgType{
		SREQ_ZERO:                     m.ZERO_CMD_RESP,
		SREQ_TARE:                     m.TARE_CMD_RESP,
		SREQ_GET_WEIGHT:               m.WEIGHT_DATA_RESP,
		SREQ_SEND_WT_CONT:             m.WEIGHT_DATA_RESP,
		SREQ_STOP_SEND_WT:             m.WEIGHT_DATA_RESP,
		SREQ_REG_WEIGHT_DATA:          m.REG_WEIGHT_RESP,
		SREQ_UNREG_WEIGHT_DATA:        m.UNREG_WEIGHT_RESP,
		SREQ_GET_RECS:                 m.GET_RECS_RESP,
		SREQ_ADD_REC:                  m.ADD_REC_RESP,
		SREQ_DEL_REC:                  m.DEL_REC_RESP,
		SREQ_DOWN_PRN_FMT:             m.DOWN_PRN_FMT_RESP,
		SREQ_GET_AP_LIST:              m.GET_AP_LIST_RESP,
		SREQ_RESCAN_AP_LIST:           m.RESCAN_AP_LIST_RESP,
		SREQ_CONNECT_AP:               m.CONNECT_AP_RESP,
		SREQ_CONNECT_AP_ONE_KEY:       m.CONNECT_AP_ONE_KEY_RESP,
		SREQ_SET_WIFI_DYNAMIC_IP:      m.SET_WIFI_DYNAMIC_IP_RESP,
		SREQ_SET_WIFI_STATIC_IP:       m.SET_WIFI_STATIC_IP_RESP,
		SREQ_GET_IP_INFO:              m.GET_IP_INFO_RESP,
		SREQ_MODIFY_BT_NAME:           m.MODIFY_BT_NAME_RESP,
		SREQ_SEND_DATA_TO_BT:          m.SEND_DATA_TO_BT_RESP,
		SREQ_SEND_DATA_TO_WIFI:        m.SEND_DATA_TO_WIFI_RESP,
		SREQ_GET_IP_MODE:              m.GET_IP_MODE_RESP, //FLF
		SREQ_GET_WIFI_INFO:            m.GET_IP_INFO_RESP,
		SREQ_GET_BUILD_INFO:           m.GET_BUILD_INFO_RESP,
		SREQ_GET_SCALE_TIME:           m.GET_SCALE_TIME_RESP,
		SREQ_SET_SCALE_TIME:           m.SET_SCALE_TIME_RESP, //20240125@FLF
		SREQ_GET_ONE_EEPROM_INFO:      m.GET_ONE_EEPROM_INFO_RESP,
		SREQ_GET_ALL_EEPROM_INFO:      m.GET_ALL_EEPROM_INFO_RESP,
		SREQ_SET_OUTPUT_FMT:           m.SET_OUTPUT_FMT_RESP,
		SREQ_OPNE_SCALE_PASSTHROUGH:   m.OPEN_SCALE_PASSTHROUGH_RESP, //20231023@FLF
		SREQ_CLOSE_SCALE_PASSTHROUGH:  m.CLOSE_SCALE_PASSTHROUGH_RESP,
		SREQ_CHANGE_SCALE_PASSTH_MODE: m.CHANGE_SCALE_PASSTH_MODE_RESP,
		SREQ_GET_SCALE_INFO:           m.GET_SCALE_INFO_RESP,
		SREQ_GET_FACTORY_INFO:         m.GET_FACTORY_INFO_RESP,
		SREQ_GET_WEIGHT_ERR:           m.GET_WEIGHT_ERR_RESP,
		SREQ_DOWN_PLU:                 m.DOWN_PLU_RESP,
		SREQ_DEL_PLU:                  m.DEL_PLU_RESP,
		SREQ_INSERT_PLU:               m.INSERT_PLU_RESP,
		SREQ_CHANGE_WIFI_MODE:         m.CHANGE_WIFI_MODE_RESP,
		SREQ_MODIFY_VAR_VALUE:         m.MODIFY_VAR_RESP,
		SREQ_SET_SERVER_IP:            m.SET_SERVER_IP_RESP,
		SREQ_EN_FACTORY_MODE:          m.EN_FACTORY_MODE_RESP,
	}
}
