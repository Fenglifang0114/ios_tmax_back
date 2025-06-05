package svc

import (
	"fmt"
	"log"
	"math"
	"math/big"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"

	"tmaxsrv/comm" // for the message types.  It is not a direct part of the code.  It is a "hel
	"tmaxsrv/lic"
	l "tmaxsrv/log"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type SrvMgr struct {
	scaleMgr *ScaleMgr
	// Registered clients.
	clients map[*Client]bool
	// Inbound messages from the clients.
	recvWsClientMsg chan []byte
	// Register requests from the clients.
	register chan *Client
	// Unregister requests from clients.
	unregister chan *Client
	// add a new scale
	addScale chan *Scale
	// remove a scale
	removeScale chan *Scale
	// Inbound messages from the scales.
	recvScaleMsg chan *ScaleRespMsg
	// Inbound messages from the scale manager
	recvScaleMgrMsg    chan *ScaleMgrRespMsg
	recvScaleMgrMsgSrv chan *SrvMgrRespMsg
	// Inbound messages from the scale's notification
	recvScaleNotifyMsg chan ScaleRespMsg
	// ScaleId to Scale map, used to access the scale
	scales map[int64]*Scale
	// scale to client map, used to send message from scale to the associated client
	clientOfScales map[*Scale]*Client
	//连接的服务器
	clientOfService map[int64]*Client
	// quitch channel to close this application
	quitch chan bool
	// productPb
	productPd *ProductRecProvider
	userPd    *UserRecProvider
	wifiPd    *WifiRecProvider
	// formulaPd
	formulaPd   *FormulaRecProvider
	flowRatePd  *FlowRateProvider
	uiConfig    *UiConfig
	modeSetting *ModeSettingProvider
	//服务与秤的关系
	srvScaleRel []*SrvScaleRel
}

var SrvIdList []int64 = []int64{999999999}

const (
	SRV_STATUS_UNINSTALLED = "status1" //服务未安装
	SRV_STATUS_INSTALLED   = "status2" //服务已安装  服务未启动
	SRV_STATUS_STARTED     = "status3" //服务已安装 服务已启动
)

type LicenseInfo struct {
	Id         string
	ModuleName string
	IsValid    bool
	ValidDate  string
}

var (
	gIsKeyValid      bool
	gMachineId       string
	gLicValidDate    string
	gModuleName      string
	gLicenseInfoList []LicenseInfo
)

func NewSrvMgr(scaleMgr *ScaleMgr, quitch chan bool) *SrvMgr {
	productPb := NewProductRecProvider()
	userPb := NewUserRecProvider()
	modeSettingPb := NewModeSettingProvider()
	wifiPb := NewWifiRecProvider()
	formulaPb := NewFormulaRecProvider()
	flowRatePb := NewFlowRateProvider()

	var licKey string

	var licKeyList []string

	licFilePath := filepath.Join(comm.GetExePath(), comm.LICENSE_FILE)

	licKey, _ = lic.ReadLicFile(licFilePath)

	licKeyList = strings.Split(licKey, "\r\n")
	for _, item := range licKeyList {
		if len(item) == 74 || len(item) == 78 {
			gIsKeyValid, gMachineId, gLicValidDate, gModuleName = lic.IsKeyValid(item)
			if gIsKeyValid {
				gLicenseInfoList = append(gLicenseInfoList, LicenseInfo{Id: gMachineId, ValidDate: gLicValidDate, ModuleName: gModuleName, IsValid: gIsKeyValid})
			}
		}
	}

	if len(gLicenseInfoList) == 0 {
		var temp []string
		licKeyList = temp
		gIsKeyValid, gMachineId, gLicValidDate, gModuleName = lic.IsKeyValid("d7a0a41239d92ee1724cd1a311ffffff2023-05-2594df26ebd828dbff03ede5f76effffff")
		gLicenseInfoList = append(gLicenseInfoList, LicenseInfo{Id: gMachineId, ValidDate: gLicValidDate, ModuleName: gModuleName, IsValid: gIsKeyValid})

	}

	return &SrvMgr{
		scaleMgr:           scaleMgr,
		scales:             map[int64]*Scale{},
		clientOfScales:     map[*Scale]*Client{},
		clientOfService:    map[int64]*Client{},
		recvWsClientMsg:    make(chan []byte),
		register:           make(chan *Client),
		unregister:         make(chan *Client),
		addScale:           make(chan *Scale, 2),
		removeScale:        make(chan *Scale, 2),
		recvScaleMsg:       make(chan *ScaleRespMsg, 2),
		recvScaleMgrMsg:    make(chan *ScaleMgrRespMsg, 2),
		recvScaleMgrMsgSrv: make(chan *SrvMgrRespMsg, 2),
		recvScaleNotifyMsg: make(chan ScaleRespMsg, 2),
		clients:            make(map[*Client]bool),
		quitch:             quitch,
		productPd:          productPb,
		userPd:             userPb,
		wifiPd:             wifiPb,
		formulaPd:          formulaPb,
		flowRatePd:         flowRatePb,
		uiConfig:           NewUiConfig(),
		modeSetting:        modeSettingPb,
		srvScaleRel:        make([]*SrvScaleRel, 0),
	}
}

func (h *SrvMgr) Run() {
	for {
		select {
		case client := <-h.register:
			scaleId := client.scaleId
			if scaleId > 0 && scaleId < 999999900 { // scaleId 0 is for management
				scale := h.scales[scaleId]
				if scale == nil {
					l.Log.Errorf("The scale: %v is not existed", scaleId)
					break
				}
			}
			isRegisted := false
			for client := range h.clients {
				if client.scaleId == scaleId {
					l.Log.Errorf("The scale: %v is already registered", scaleId)
					isRegisted = true
					break
				}
			}
			if !isRegisted {
				h.clients[client] = true
				// if scaleId != 0 { // 0 reserved for common information channel, 9999 reserved for legacy MCU scale, only support one scale with this ID
				if scaleId < 999999900 {
					h.clientOfScales[h.scales[scaleId]] = client
				}
				if scaleId != 0 && scaleId < 999999900 { // id 0 is reserved for common information channel
					h.scales[scaleId].SetClient(client)
				}

				if scaleId > 999999900 { // id 0 is reserved for common information channel
					h.clientOfService[scaleId] = client
				}

				// h.clientOfScales[h.scales[scaleId]] = client
				//}
			} else {
				// close(client.send) // either other client is already registered the scale or the scale is not yet registered
			}
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				// TODO: handle client disconnect
				//0 通道也断了，如何处理，先将下面的程序改为0的话断掉也直接断开。
				if client.scaleId != -1 { //20250220    -1 原来是 0
					if h.scales[client.scaleId] != nil {
						h.scales[client.scaleId].HandleClientDisconnect()
					}

				}
				// client.Close()
				delete(h.clients, client)
				if client.scaleId > 999999900 {
					delete(h.clientOfService, client.scaleId)
				} else {
					delete(h.clientOfScales, h.scales[client.scaleId])
				}

				// close(client.send)
			}
		case scale := <-h.addScale: // from scale manager
			if len(h.scales) == 0 {
				h.scales[scale.Id] = scale
			} else {
				if h.scales[scale.Id] == nil {
					h.scales[scale.Id] = scale
				} else {
					l.Log.Warnf("scale id already registered, just ignore it")
				}
			}
		case scale := <-h.removeScale: // from scale manager
			if h.scales[scale.Id] != nil { // scale not existing
				h.scales[scale.Id] = nil
			} else {
				l.Log.Warn("scale id not registered and can't be removed, just ignore it")
			}
		case userMessage := <-h.recvWsClientMsg: // message from websocket client
			var data map[string][]byte
			json.Unmarshal(userMessage, &data)
			scaleId := new(big.Int).SetBytes(data["scaleId"]).Int64()
			if scaleId == 0 { // not for scale communication but for information purposes
				// request := parseMsg(string(data["message"]))
				// TODO: send request to scale manager to get scale list or get serial ports
				// parse request for "get port list", "get scale list", "update scale conneciton",
				//                   "create a new scale", delete a scale" or "close application"
				l.Log.Infof("Got request from common channel %v\n", string(data["message"]))
				parseMsgAndTrigEvt(h.scaleMgr, string(data["message"]))

			} else if scaleId > 999999900 { // not for scale communication but for information purposes
				//小服务的接口
				l.Log.Infof("Got request from common channel %v\n", string(data["message"]))
				parseMsgAndTrigEvtService(h.scaleMgr, string(data["message"]), scaleId)
			} else {
				scale := h.scales[scaleId]
				if scale == nil { // something wrong about scale id
					l.Log.Warnf("cannot find the scale with id: %v\n", scaleId)
				} else if scale.Id != scaleId { // something wrong about scale id
					l.Log.Errorf("scale id: %v is not consistent with the id: %v recorded in the hub", scaleId, scale.Id)
				} else {
					// send data to the scale
					// if req, err := parseToScaleReq(string(data["message"])); err == nil {
					// 	go procToScaleReq(req, scaleId, h, scale) // TODO: handle error
					// }

				}
			}
		case scaleMessage := <-h.recvScaleMsg:
			// handle the message from the scale

			if h.scales[scaleMessage.ScaleId] != nil {
				client := h.clientOfScales[h.scales[scaleMessage.ScaleId]]
				if client != nil {
					outData, _ := json.Marshal(scaleMessage)
					client.sendCh <- outData
				}
			}

			// default:
			// 	fmt.Println("    .")
			// 	time.Sleep(1 * time.Millisecond)
		case scaleMgrMessage := <-h.recvScaleMgrMsg:
			// handle the message from the scale
			if len(h.clientOfScales) == 0 {
				break
			}
			if h.clientOfScales[h.scales[0]] == nil {
				break
			}

			client := h.clientOfScales[h.scales[0]]
			println(h.scales)
			outData, _ := json.Marshal(scaleMgrMessage)
			if client != nil {
				client.sendCh <- outData
				l.Log.Debugf("ClientSendch---------- %v\n", string(outData))
			}
		case recvScaleMgrMsgSrv := <-h.recvScaleMgrMsgSrv: //20241118 如何将数据传出去？9999999999  99999998
			// handle the message from the scale
			// 只要是服务，全送
			outData, _ := json.Marshal(recvScaleMgrMsgSrv)
			if recvScaleMgrMsgSrv.ScaleId > 999999900 {
				client := h.clientOfService[recvScaleMgrMsgSrv.ScaleId]
				if client != nil {
					client.sendCh <- outData
				}

			} else {

				for _, srvRel := range h.srvScaleRel {
					if srvRel.ScaleId == recvScaleMgrMsgSrv.ScaleId && srvRel.IsUsed {
						client := h.clientOfService[srvRel.SrvId]
						if client != nil {
							client.sendCh <- outData
						}

					}
				}

			}

		case scaleMessage := <-h.recvScaleNotifyMsg:
			// handle the message from the scale
			if len(h.clientOfScales) == 0 {
				break
			}

			if h.clientOfScales[h.scales[scaleMessage.ScaleId]] == nil {
				break
			}

			client := h.clientOfScales[h.scales[scaleMessage.ScaleId]]
			if client != nil { // handle the transient situation
				outData, _ := json.Marshal(scaleMessage)
				// fmt.Printf("%v\n", scaleMessage)
				l.Log.Debugf("%v\n", string(outData))
				client.sendCh <- outData
			}
		default:
			time.Sleep(time.Microsecond * 100)
			continue
		}
	}
}

func parseMsgAndTrigEvt(scaleMgr *ScaleMgr, reqJson string) {
	var req Request
	if err := json.UnmarshalFromString(reqJson, &req); err != nil {
		l.Log.Error(err)
		return
	}

	switch req.Req {
	case REQ_GET_SCALE_LIST:
		ScaleListed.Trigger(scalesListed, scaleMgr)
	case REQ_GET_PORT_LIST:
		PortsListed.Trigger(portsListed)
	case REQ_ADD_SCALE:
		jsonStr := req.ReqData
		var data ReqAddScale
		if err := json.UnmarshalFromString(jsonStr, &data); err != nil {
			l.Log.Error(err)
		} else {
			ScaleAdded.Trigger(scaleAdded, data)
		}
	case REQ_DEL_SCALE:
		jsonStr := req.ReqData
		var data ReqDelScale
		if err := json.UnmarshalFromString(jsonStr, &data); err != nil {
			l.Log.Error(err)
		} else {
			ScaleDeleted.Trigger(scaleDeleted, data)
		}
	case REQ_MODIFY_SCALE:
		jsonStr := req.ReqData
		var data ReqModifyScale
		if err := json.UnmarshalFromString(jsonStr, &data); err != nil {
			l.Log.Error(err)
		} else {
			ScaleModified.Trigger(scaleModified, data)
		}
	case REQ_MODIFY_SCALE_NAME:
		jsonStr := req.ReqData
		var data ReqModifyScaleName
		if err := json.UnmarshalFromString(jsonStr, &data); err != nil {
			l.Log.Error(err)
		} else {
			ScaleNameModified.Trigger(scaleNameModified, data)
		}

	case REQ_GET_PRODUCT_LIST:
		productsListed.Trigger(scaleMgr.srvMgr)
	case REQ_ADD_PRODUCT:
		jsonStr := req.ReqData
		var data ReqAddProductList
		if err := json.UnmarshalFromString(jsonStr, &data); err != nil {
			l.Log.Error(err)
		} else {
			ProductAdded.Trigger(productAdded, scaleMgr.srvMgr, data)
		}
	case REQ_DEL_PRODUCT:
		jsonStr := req.ReqData
		var data ReqDelProduct
		if err := json.UnmarshalFromString(jsonStr, &data); err != nil {
			l.Log.Error(err)
		} else {
			ProductDeleted.Trigger(productDeleted, scaleMgr.srvMgr, data)
		}
	case REQ_DEL_ALL_PRODUCT:
		productDeletedAll.Trigger(scaleMgr.srvMgr)
	case REQ_MODIFY_PRODUCT:
		jsonStr := req.ReqData
		var data ReqAddProductList
		if err := json.UnmarshalFromString(jsonStr, &data); err != nil {
			l.Log.Error(err)
		} else {
			ProductModified.Trigger(productModified, scaleMgr.srvMgr, data)
		}
	case REQ_GET_USER_LIST:
		usersListed.Trigger(scaleMgr.srvMgr)
	case REQ_ADD_USER:
		jsonStr := req.ReqData
		var data ReqAddUser
		if err := json.UnmarshalFromString(jsonStr, &data); err != nil {
			l.Log.Error(err)
		} else {
			UserAdded.Trigger(userAdded, scaleMgr.srvMgr, data)
		}
	case REQ_DEL_USER:
		jsonStr := req.ReqData
		var data ReqDelUser
		if err := json.UnmarshalFromString(jsonStr, &data); err != nil {
			l.Log.Error(err)
		} else {
			UserDeleted.Trigger(userDeleted, scaleMgr.srvMgr, data)
		}
	case REQ_MODIFY_USER:
		jsonStr := req.ReqData
		var data ReqModifyUser
		if err := json.UnmarshalFromString(jsonStr, &data); err != nil {
			l.Log.Error(err)
		} else {
			UserModified.Trigger(userModified, scaleMgr.srvMgr, data)
		}

	case REQ_QUIT_APPLICATION:
		l.Log.Warn("Got quit application")
		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_QUIT_APPLICATION, MsgBody: ""}
		mSrvMgr.quitch <- true
	// case REQ_GET_UI_CONF:
	// 	l.Log.Info("Got get UI Config request")
	// 	config, _ := mSrvMgr.uiConfig.GetConfig()
	// 	configStr, _ := json.MarshalToString(config)
	// 	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_GET_UI_CONFIG, MsgBody: configStr}
	// case REQ_UPDATE_UI_CONF:
	// 	l.Log.Info("Got update UI Config request")
	// 	var config Config
	// 	if err := json.UnmarshalFromString(req.ReqData, &config); err != nil {
	// 		l.Log.Error(err)
	// 		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_UPDATE_UI_CONFIG, MsgBody: "failed to parse update UI Config"}
	// 	} else {
	// 		err := mSrvMgr.uiConfig.UpdateConfig(&config)
	// 		if err != nil {
	// 			l.Log.Error(err)
	// 		}
	// 		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_UPDATE_UI_CONFIG, MsgBody: ""}
	// 	}
	case REQ_GET_LICENSE:
		getLicenseList()
		licListStr, _ := json.MarshalToString(gLicenseInfoList)
		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_GET_LICENSE, MsgBody: licListStr}

	case REQ_CHECK_LICENSE_KEY:

		isValid, machineId, licValidDate, moduleName := lic.IsKeyValid(req.ReqData)
		var isValidStr string
		if isValid {
			isValidStr = "true"
		} else {
			isValidStr = "false"
		}
		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_CHECK_LICENSE_KEY, MsgBody: moduleName + "," + isValidStr + "," + machineId + "," + licValidDate}

	case REQ_UPDATE_LICENSE:
		licPath := filepath.Join(comm.GetExePath(), comm.LICENSE_FILE)
		if err := lic.SaveKey(licPath, req.ReqData); err != nil {
			mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_UPDATE_LICENSE, MsgBody: "fail"}
		}
		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_UPDATE_LICENSE, MsgBody: "ok"}

	case REQ_GET_DETAIL_LIST: // TODO: should we check the input parameters?
		DetailListed.Trigger(detailListed, scaleMgr)
	case REQ_GET_SCALE_SRV_LIST: // TODO: should we check the input parameters?
		jsonStr := req.ReqData
		scaleSrvList.Trigger(scaleMgr, jsonStr)
	case REQ_SET_SCALE_SRV_VAL: // TODO: should we check the input parameters?
		jsonStr := req.ReqData
		var data SrvScaleRel
		if err := json.UnmarshalFromString(jsonStr, &data); err != nil {
			l.Log.Error(err)
		} else {
			setScaleSrvVal.Trigger(scaleMgr, data)
		}

	case REQ_GET_WIFI_PWD_LIST:
		wifiListed.Trigger(scaleMgr.srvMgr)
	case REQ_WIFI_PWD:
		jsonStr := req.ReqData
		var data ReqAddWifi
		if err := json.UnmarshalFromString(jsonStr, &data); err != nil {
			l.Log.Error(err)
		} else {
			wifiAdded.Trigger(scaleMgr.srvMgr, data)
		}
	case REQ_SEND_TO_SRV1:
		jsonStr := req.ReqData
		sendToSrv1.Trigger(scaleMgr.srvMgr, jsonStr)

	case REQ_SET_DO_SERVICE_ACTION:
		jsonStr := req.ReqData
		var data ReqDoServiceAction
		if err := json.UnmarshalFromString(jsonStr, &data); err != nil {
			l.Log.Error(err)
		} else {
			doServiceAction.Trigger(scaleMgr.srvMgr, data)
		}
		//新增原料类型
	case REQ_ADD_RAW_TYPE:
		jsonStr := req.ReqData
		var data ReqAddRawType
		if err := json.UnmarshalFromString(jsonStr, &data); err != nil {
			l.Log.Error(err)
		} else {
			RawTypeAdded.Trigger(rawTypeAdded, scaleMgr.srvMgr, data)
		}

		//新增配方类型
	case REQ_ADD_FORMULA_TYPE:
		jsonStr := req.ReqData
		var data ReqAddFormulaType
		if err := json.UnmarshalFromString(jsonStr, &data); err != nil {
			l.Log.Error(err)
		} else {
			FormulaTypeAdded.Trigger(formulaTypeAdded, scaleMgr.srvMgr, data)
		}
	case REQ_GET_FORMULA_TYPE_LIST:
		formulaTypeListed.Trigger(scaleMgr.srvMgr)
		//获取原料类型表
	case REQ_GET_RAW_TYPE_LIST:
		rawTypeListed.Trigger(scaleMgr.srvMgr)
		//新增原料数据
	case REQ_ADD_RAW_DATA:
		jsonStr := req.ReqData
		var data ReqAddRawData
		if err := json.UnmarshalFromString(jsonStr, &data); err != nil {
			l.Log.Error(err)
		} else {
			RawDataAdded.Trigger(rawDataAdded, scaleMgr.srvMgr, data)
		}
	case REQ_GET_RAW_DATA_LIST:
		rawDataListed.Trigger(scaleMgr.srvMgr)
	case REQ_EDIT_RAW_DATA:
		jsonStr := req.ReqData
		var data ReqEditRawData
		if err := json.UnmarshalFromString(jsonStr, &data); err != nil {
			l.Log.Error(err)
		} else {
			RawDataEdited.Trigger(rawDataEdited, scaleMgr.srvMgr, data)
		}
	case REQ_DELETE_RAW_DATA:
		jsonStr := req.ReqData
		var data ReqDelRawData
		if err := json.UnmarshalFromString(jsonStr, &data); err != nil {
			l.Log.Error(err)
		}
		rawDataDeleted.Trigger(scaleMgr.srvMgr, data)

	//新增配方
	case REQ_ADD_FORMULA_DATA:
		jsonStr := req.ReqData
		var data ReqAddFormulaData
		if err := json.UnmarshalFromString(jsonStr, &data); err != nil {
			l.Log.Error(err)
		} else {
			FormulaDataAdded.Trigger(formulaDataAdded, scaleMgr.srvMgr, data)
		}
	//修改配方
	case REQ_EDIT_FORMULA_DATA:
		jsonStr := req.ReqData
		var data ReqAddFormulaData
		if err := json.UnmarshalFromString(jsonStr, &data); err != nil {
			l.Log.Error(err)
		} else {
			formulaDataEdited.Trigger(scaleMgr.srvMgr, data)
		}
	case REQ_GET_FORMULA_LIST:
		formulaRecList.Trigger(scaleMgr.srvMgr)

		//新增配方称重记录
	case REQ_ADD_FORMULA_REC:
		jsonStr := req.ReqData
		var data ReqFormulaWgtRec
		if err := json.UnmarshalFromString(jsonStr, &data); err != nil {
			l.Log.Error(err)
		} else {
			FormulaWgtRecAdded.Trigger(formulaWgtRecAdded, scaleMgr.srvMgr, data)
		}
		//获取配方称重记录
	case REQ_GET_FORMULA_REC_LIST:
		formulaWgtRecList.Trigger(scaleMgr.srvMgr)
		//删除配方
	case REQ_DELETE_FORMULA_DATA:
		jsonStr := req.ReqData
		var data ReqDelFmaData
		if err := json.UnmarshalFromString(jsonStr, &data); err != nil {
			l.Log.Error(err)
		} else {
			FormulaDeleted.Trigger(formulaDeleted, scaleMgr.srvMgr, data)
		}
		//新增流速
	case REQ_ADD_FLOW_RATE:
		jsonStr := req.ReqData
		var data ReqFlowRateRec
		if err := json.UnmarshalFromString(jsonStr, &data); err != nil {
			l.Log.Error(err)
		} else {
			FlowRateAdded.Trigger(flowRateAdded, scaleMgr.srvMgr, data)
		}
		//获取流速
	case REQ_GET_FLOW_RATE_LIST:
		flowRateList.Trigger(scaleMgr.srvMgr)

	}

}

func getLicenseList() {
	var licKey string
	var newLicList []LicenseInfo
	var licKeyList []string

	licFilePath := filepath.Join(comm.GetExePath(), comm.LICENSE_FILE)

	licKey, _ = lic.ReadLicFile(licFilePath)

	licKeyList = strings.Split(licKey, "\r\n")
	for _, item := range licKeyList {
		if len(item) == 74 || len(item) == 78 {
			gIsKeyValid, gMachineId, gLicValidDate, gModuleName = lic.IsKeyValid(item)
			if gIsKeyValid {
				newLicList = append(newLicList, LicenseInfo{Id: gMachineId, ValidDate: gLicValidDate, ModuleName: gModuleName, IsValid: gIsKeyValid})
			}
		}
	}

	if len(newLicList) == 0 {
		var temp []string
		licKeyList = temp
		gIsKeyValid, gMachineId, gLicValidDate, gModuleName = lic.IsKeyValid("d7a0a41239d92ee1724cd1a311ffffff2023-05-2594df26ebd828dbff03ede5f76effffff")
		newLicList = append(newLicList, LicenseInfo{Id: gMachineId, ValidDate: gLicValidDate, ModuleName: gModuleName, IsValid: gIsKeyValid})

	}
	gLicenseInfoList = newLicList
}

func parseMsgAndTrigEvtService(scaleMgr *ScaleMgr, reqJson string, scaleId int64) {
	var req Request
	if err := json.UnmarshalFromString(reqJson, &req); err != nil {
		l.Log.Error(err)
		return
	}

	switch req.Req {
	case REQ_GET_SCALE_LIST:
		scalesListedSrv.Trigger(scaleMgr, scaleId)
	case REQ_SEND_TO_UI:
		jsonStr := req.ReqData
		println(jsonStr)
		println("--------------------")
		sendToUi.Trigger(scaleMgr.srvMgr, jsonStr)

	}
}

func (p productListedNotifier) Handle(mgr *SrvMgr) {
	// Do something for this event
	l.Log.Debug("Handle productListedNotifier called")
	// Do something with this event

	products, _ := mgr.productPd.GetRecsList()

	batchSize := 100 //每次发送1000条
	numBatches := (len(products) + batchSize - 1) / batchSize

	for i := 0; i < numBatches; i++ {
		startIndex := i * batchSize
		endIndex := (i + 1) * batchSize
		if endIndex > len(products) {
			endIndex = len(products)
		}
		batchProducts := products[startIndex:endIndex]
		var productsStr string
		var err error
		if productsStr, err = json.MarshalToString(batchProducts); err != nil {
			l.Log.Error(err)
			// TODO: error handling
		}
		// 发送每一批次的数据
		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_PRODUCTS_LIST, MsgBody: productsStr}
		time.Sleep(100 * time.Millisecond)
	}
}

func (p addProductNotifier) Handle(mgr *SrvMgr, payload ReqAddProductList) {
	// Do something for this event
	l.Log.Debug("Handle addProductNotifier called")
	// rec := ProductRec{Id: payload.Id, Product: payload.Product, WithPretare: payload.WithPretare, Pretare: payload.Pretare, Remarks: payload.Remarks}
	// if err := mgr.productPd.InsertRec(rec); err != nil {
	// 	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_PRODUCT_ADD, MsgBody: err.Error()}
	// }
	var recList []ProductRec
	for _, product := range payload {
		rec := ProductRec{
			Plu:         product.Plu,
			ProductCode: product.ProductCode,
			ItemCode:    product.ItemCode,
			Category:    product.Category,
			ProductName: product.ProductName,
			GeneralUnit: product.GeneralUnit,
			TaxType:     product.TaxType,
			Price:       product.Price,
			UnitWeight:  product.UnitWeight,
			Pretare:     product.Pretare,
			LimitHigh:   product.LimitHigh,
			LimitLow:    product.LimitLow,
		}
		recList = append(recList, rec)

		// if err := mgr.productPd.InsertRec(rec); err != nil {
		// 	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_PRODUCT_ADD, MsgBody: err.Error()}
		// 	return
		// }
	}
	if err := mgr.productPd.Insert100Rec(recList); err != nil {
		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_PRODUCT_ADD, MsgBody: err.Error()}
		return
	}
	// send result back to requestee
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_PRODUCT_ADD, MsgBody: "OK"}
}

func (p delProductNotifier) Handle(mgr *SrvMgr, payload ReqDelProduct) {
	// Do something for this event
	l.Log.Debug("Handle delProductNotifier called")
	if err := NewProductRecProvider().DeleteRec(uint(payload.RecId)); err != nil {
		// TODO: error handling
	}
	// send ports list back to requestee
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_PRODUCT_DEL, MsgBody: ""}
}

func (p delAllProductNotifier) Handle(mgr *SrvMgr) {
	// Do something for this event
	res := "OK"
	l.Log.Debug("Handle delProductNotifier called")
	if err := NewProductRecProvider().DeleteAllRec(); err != nil {
		res = ""
	}
	// send ports list back to requestee
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_PRODUCT_DEL, MsgBody: res}
}

func (p modifyProductNotifier) Handle(mgr *SrvMgr, payload ReqAddProductList) {

	var recList []ProductRec
	for _, product := range payload {
		rec := ProductRec{
			Plu:         product.Plu,
			ProductCode: product.ProductCode,
			ItemCode:    product.ItemCode,
			Category:    product.Category,
			ProductName: product.ProductName,
			GeneralUnit: product.GeneralUnit,
			TaxType:     product.TaxType,
			Price:       product.Price,
			UnitWeight:  product.UnitWeight,
			Pretare:     product.Pretare,
			LimitHigh:   product.LimitHigh,
			LimitLow:    product.LimitLow,
		}
		recList = append(recList, rec)
	}

	if err := mgr.productPd.BatchModifyRec(recList); err != nil {
		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_PRODUCT_MODIFY, MsgBody: err.Error()}
		return
	}

	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_PRODUCT_MODIFY, MsgBody: "OK"}

}

func (p userListedNotifier) Handle(mgr *SrvMgr) {
	// Do something for this event
	l.Log.Debug("Handle userListedNotifier called")
	// Do something with this event
	users, _ := NewUserRecProvider().GetRecsList()

	var userStr string
	var err error
	if userStr, err = json.MarshalToString(users); err != nil {
		l.Log.Error(err)
		// TODO: error handling
	}
	// send users list back to requestee
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_USERS_LIST, MsgBody: userStr}
}

func (p addUserNotifier) Handle(mgr *SrvMgr, payload ReqAddUser) {
	// Do something for this event
	l.Log.Debug("Handle addUserNotifier called")
	var rec UserRec = UserRec{Id: payload.Id, Name: payload.Name, Phone: payload.Phone, IsFemale: payload.IsFemale, Remarks: payload.Remarks}
	if err := NewUserRecProvider().InsertRec(rec); err != nil {
		// TODO: error handling
		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_USER_ADD, MsgBody: err.Error()}
	}
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_USER_ADD, MsgBody: ""}
}

func (p delUserNotifier) Handle(mgr *SrvMgr, payload ReqDelUser) {
	// Do something for this event
	l.Log.Debug("Handle delUserNotifier called")
	if err := NewUserRecProvider().DeleteRec(uint(payload.RecId)); err != nil {
		// TODO: error handling
		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_USER_DEL, MsgBody: err.Error()}
	}
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_USER_DEL, MsgBody: ""}
}

func (p modifyUserNotifier) Handle(mgr *SrvMgr, payload ReqModifyUser) {
	// Do something for this event
	l.Log.Debug("Handle modifyUserNotifier called")
	var rec UserRec = UserRec{RecId: uint(payload.RecId), Id: payload.Id, Name: payload.Name, Phone: payload.Phone, IsFemale: payload.IsFemale, Remarks: payload.Remarks}
	if err := mgr.userPd.ModifyRec(rec); err != nil {
		// TODO: error handling
		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_USER_MODIFY, MsgBody: err.Error()}
	}
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_USER_MODIFY, MsgBody: ""}
}

func (p wifiPwdListedNotifier) Handle(mgr *SrvMgr) {
	// Do something for this event
	l.Log.Debug("Handle wifiPwdListedNotifier called")
	// Do something with this event
	wifis, _ := NewWifiRecProvider().GetRecsList()

	var wifiStr string
	var err error
	if wifiStr, err = json.MarshalToString(wifis); err != nil {
		l.Log.Error(err)
		// TODO: error handling
	}
	// send wifi pwd list back to requestee
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_WIFI_PWD_LIST, MsgBody: wifiStr}
}

func (p addWifiPwdNotifier) Handle(mgr *SrvMgr, payload ReqAddWifi) {
	// Do something for this event
	l.Log.Debug("Handle addWifiNotifier called")
	var rec WifiRec = WifiRec{Ssid: payload.Ssid, Pwd: payload.Pwd}
	if err := NewWifiRecProvider().InsertRec(rec); err != nil {
		// TODO: error handling
		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_WIFI_PWD_ADD, MsgBody: err.Error()}
	}
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_WIFI_PWD_ADD, MsgBody: "ok"}
}

func (p doServiceActionNotifier) Handle(mgr *SrvMgr, payload ReqDoServiceAction) {
	// Do something for this event
	l.Log.Debug("Handle addWifiNotifier called")
	msgStr := ""
	// if mgr.clientOfService[payload.ServiceId] != nil {
	// 	msgStr = "Service Started"
	// 	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_DO_SERVICE_ACTION, MsgBody: msgStr}
	// 	return
	// }

	srvName, srvPath := getServiceNameAndPath(payload.ServiceId)
	if srvPath == "" {
		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_DO_SERVICE_ACTION, MsgBody: msgStr}
		return
	}

	serviceName := srvName
	servicePath := srvPath
	serviceManager := ServiceManager{
		ServiceName: serviceName,
		ServicePath: servicePath,
	}

	switch payload.Action {
	case "Install":
		if err := serviceManager.Install(); err != nil {
			// TODO: error handling
			mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_DO_SERVICE_ACTION, MsgBody: err.Error()}
			return
		}
		msgStr = SRV_STATUS_INSTALLED

	case "Start":
		if err := serviceManager.Start(); err != nil {
			// TODO: error handling
			mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_DO_SERVICE_ACTION, MsgBody: err.Error()}
			return
		}
		msgStr = SRV_STATUS_STARTED
	case "Stop":
		if err := serviceManager.Stop(); err != nil {
			// TODO: error handling
			mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_DO_SERVICE_ACTION, MsgBody: err.Error()}
			return
		}
		msgStr = SRV_STATUS_INSTALLED
	case "Uninstall":
		if err := serviceManager.Uninstall(); err != nil {
			// TODO: error handling
			mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_DO_SERVICE_ACTION, MsgBody: err.Error()}
			return
		}
		msgStr = SRV_STATUS_UNINSTALLED

	case "Status":
		status := getServiceStatus(serviceName)
		msgStr = "Status:" + status
	default:
		break

	}

	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_DO_SERVICE_ACTION, MsgBody: msgStr}
}

func getServiceStatus(serviceName string) string {
	res := IsServiceInstalled(serviceName)
	if !res {
		//服务没有安装
		return SRV_STATUS_UNINSTALLED
	} else if !IsServiceRunning(serviceName) {
		return SRV_STATUS_INSTALLED
	} else {
		return SRV_STATUS_STARTED
	}
}

func getServiceNameAndPath(srvId int64) (string, string) {
	switch srvId {
	case 999999999:
		return "RetailDetailService", filepath.Join(comm.GetServicePath(), "detailservice.exe")
	default:
		return "", ""
	}
}

// ServiceManager结构体用于管理服务的操作
type ServiceManager struct {
	ServiceName string
	ServicePath string
}

// Install方法用于安装服务
func (sm *ServiceManager) Install() error {
	installCmd := fmt.Sprintf("sc create %s binPath= \"%s\"  start= auto", sm.ServiceName, sm.ServicePath)
	// installCmd := "sc create RetailDetailService binpath=\"G:\\T-max\\wifi_tmax\\service\\TmaxService\\Backend\\srvdata\\service\\detailservice.exe\""

	cmd := exec.Command("cmd", "/C", installCmd)

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("fail: %v", err)
	}
	fmt.Println("ok")
	return nil
}

// Uninstall方法用于卸载服务
func (sm *ServiceManager) Uninstall() error {
	uninstallCmd := fmt.Sprintf("sc delete %s", sm.ServiceName)
	cmd := exec.Command("cmd", "/C", uninstallCmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("fail: %v", err)
	}
	fmt.Println("ok")
	return nil
}

// Start方法用于启动服务
func (sm *ServiceManager) Start() error {
	startCmd := fmt.Sprintf("sc start %s", sm.ServiceName)
	cmd := exec.Command("cmd", "/C", startCmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("fail: %v", err)
	}
	fmt.Println("ok")
	return nil
}

// Stop方法用于停止服务
func (sm *ServiceManager) Stop() error {
	stopCmd := fmt.Sprintf("sc stop %s", sm.ServiceName)
	cmd := exec.Command("cmd", "/C", stopCmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("fail: %v", err)
	}
	fmt.Println("ok")
	return nil
}

func isServiceRunning(serviceName string) (bool, error) {
	m, err := mgr.Connect()
	if err != nil {
		return false, err
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return false, err
	}
	defer s.Close()

	status, err := s.Query()
	if err != nil {
		return false, err
	}

	return status.State == svc.Running, nil
}

func IsServiceInstalled(serviceName string) bool {
	m, err := mgr.Connect()
	if err != nil {
		fmt.Errorf("%v", err)
	}
	defer m.Disconnect()

	services, err := m.ListServices()
	if err != nil {
		fmt.Errorf("%v", err)
	}

	for _, s := range services {
		if s == serviceName {
			return true
		}
	}
	return false
}

func IsServiceRunning(serviceName string) bool {
	m, err := mgr.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	status, err := s.Query()
	if err != nil {
		log.Fatal(err)
	}

	return status.State == svc.Running
}

//配方秤

func (p addRawTypeNotifier) Handle(mgr *SrvMgr, payload ReqAddRawType) {
	// Do something for this event
	l.Log.Debug("Handle addRawTypeNotifier called")
	var rec RawMaterialCategory = RawMaterialCategory{CategoryName: payload.Name}
	if err := NewFormulaRecProvider().InsertRawType(rec); err != nil {
		// TODO: error handling
		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_RAW_TYPE_ADD, MsgBody: err.Error()}
	}
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_RAW_TYPE_ADD, MsgBody: ""}
}

func (p addFormulaTypeNotifier) Handle(mgr *SrvMgr, payload ReqAddFormulaType) {
	// Do something for this event
	l.Log.Debug("Handle addFormulaTypeNotifier called")
	var rec FormulaCategory = FormulaCategory{CategoryName: payload.Name}
	if err := NewFormulaRecProvider().InsertFormulaType(rec); err != nil {
		// TODO: error handling
		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_FORMULA_TYPE_ADD, MsgBody: err.Error()}
	}
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_FORMULA_TYPE_ADD, MsgBody: ""}
}

func (p getRawTypeListNotifier) Handle(mgr *SrvMgr) {
	// Do something for this event
	l.Log.Debug("Handle getRawTypeListNotifier called")
	types, _ := NewFormulaRecProvider().GetRawTypeList()
	var typesStr string
	var err error
	if typesStr, err = json.MarshalToString(types); err != nil {
		l.Log.Error(err)

	}
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_RAW_TYPE_LIST, MsgBody: typesStr}
}

func (p getFormulaTypeListNotifier) Handle(mgr *SrvMgr) {
	// Do something for this event
	l.Log.Debug("Handle getFormulaTypeListNotifier called")
	types, _ := NewFormulaRecProvider().GetFormulaTypeList()
	var typesStr string
	var err error
	if typesStr, err = json.MarshalToString(types); err != nil {
		l.Log.Error(err)

	}
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_FORMULA_TYPE_LIST, MsgBody: typesStr}
}

func (p rawDataAddedNotifier) Handle(mgr *SrvMgr, payload ReqAddRawData) {
	// Do something for this event
	l.Log.Debug("Handle addRawTypeNotifier called")
	var rec RawMaterial = RawMaterial{
		MaterialID:   payload.MaterialID,
		MaterialName: payload.MaterialName,
		CategoryID:   payload.CategoryID,
		Ingredient:   payload.Ingredient,
		Remark:       payload.Remark,
		Remark1:      payload.Remark1,
		CreatedBy:    payload.CreatedBy,
		UpdatedBy:    payload.UpdatedBy,
	}
	if err := NewFormulaRecProvider().InsertRawInfo(rec); err != nil {
		// TODO: error handling
		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_RAW_DATA_ADD, MsgBody: err.Error()}
	}
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_RAW_DATA_ADD, MsgBody: "ok"}
}

func (p rawDataListedNotifier) Handle(mgr *SrvMgr) {
	// Do something for this event
	l.Log.Debug("Handle rawDataListedNotifier called")
	recs, _ := NewFormulaRecProvider().GetRawDataList()

	var typesStr string
	var err error
	if typesStr, err = json.MarshalToString(recs); err != nil {
		l.Log.Error(err)

	}
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_RAW_LIST, MsgBody: typesStr}
}

func (p rawDataEditedNotifier) Handle(mgr *SrvMgr, payload ReqEditRawData) {
	// Do something for this event
	l.Log.Debug("Handle rawDataEditedNotifier called")

	var rec RawMaterial = RawMaterial{
		RecId:        payload.RecId,
		MaterialID:   payload.MaterialID,
		MaterialName: payload.MaterialName,
		CategoryID:   payload.CategoryID,
		Ingredient:   payload.Ingredient,
		Remark:       payload.Remark,
		Remark1:      payload.Remark1,
		CreatedBy:    payload.CreatedBy,
		UpdatedBy:    payload.UpdatedBy,
	}

	if err := NewFormulaRecProvider().UpdateRawInfo(rec); err != nil {
		// TODO: error handling
		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_RAW_DATA_EDIT, MsgBody: err.Error()}
	}
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_RAW_DATA_EDIT, MsgBody: "ok"}
}

func (p rawDataDeletedNotifier) Handle(mgr *SrvMgr, payload ReqDelRawData) {
	// Do something for this event
	l.Log.Debug("Handle rawDataDeletedNotifier called")
	rec := payload.RecId
	if err := NewFormulaRecProvider().DeleteRawInfo(rec); err != nil {
		// TODO: error handling
		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_RAW_DATA_DELETE, MsgBody: err.Error()}
	}
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_RAW_DATA_DELETE, MsgBody: "ok"}
}

// 增加配方
func (p addFormulaRecNotifier) Handle(mgr *SrvMgr, payload ReqAddFormulaData) {
	// Do something for this event
	l.Log.Debug("Handle addFormulaRecNotifier called")

	var header FormulaHeader = FormulaHeader{
		FormulaID:     payload.Header.FormulaID,
		FormulaName:   payload.Header.FormulaName,
		CategoryID:    payload.Header.CategoryID,
		Remark:        payload.Header.Remark,
		CreatedBy:     payload.Header.CreatedBy,
		UpdatedBy:     payload.Header.UpdatedBy,
		FormulaMode:   payload.Header.FormulaMode,
		FormulaUnit:   payload.Header.FormulaUnit,
		TotalWeight:   payload.Header.TotalWeight,
		MaterialCount: payload.Header.MaterialCount,
		IsEncrypted:   payload.Header.IsEncrypted,
		NeedContainer: payload.Header.NeedContainer,
	}

	if err := NewFormulaRecProvider().InsertFormulaHeader(header); err != nil {
		// TODO: error handling
		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_FORMULA_ADD, MsgBody: err.Error()}
	}

	headerId, _ := NewFormulaRecProvider().GetMaxFormulaRecId() //获取最新的配方ID

	for _, detail := range payload.Detail {
		tempRec := FormulaDetail{
			FormulaRecID:       headerId,
			MaterialID:         detail.MaterialID,
			MaterialWeight:     detail.MaterialWeight,
			MaterialPercentage: detail.MaterialPercentage,
			Sequence:           detail.Sequence,
			AllowableError:     detail.AllowableError,
			Remark:             detail.Remark,
		}
		if err := NewFormulaRecProvider().InsertFormulaBody(tempRec); err != nil {
			// TODO: error handling
			mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_FORMULA_ADD, MsgBody: err.Error()}
		}
	}

	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_FORMULA_ADD, MsgBody: "ok"}

}

// 修改配方
func (p editFormulaRecNotifier) Handle(mgr *SrvMgr, payload ReqAddFormulaData) {
	// Do something for this event
	l.Log.Debug("Handle editFormulaRecNotifier called")

	var header FormulaHeader = FormulaHeader{
		FormulaID:     payload.Header.FormulaID,
		FormulaName:   payload.Header.FormulaName,
		CategoryID:    payload.Header.CategoryID,
		Remark:        payload.Header.Remark,
		CreatedBy:     payload.Header.CreatedBy,
		UpdatedBy:     payload.Header.UpdatedBy,
		FormulaMode:   payload.Header.FormulaMode,
		FormulaUnit:   payload.Header.FormulaUnit,
		TotalWeight:   payload.Header.TotalWeight,
		MaterialCount: payload.Header.MaterialCount,
		IsEncrypted:   payload.Header.IsEncrypted,
		NeedContainer: payload.Header.NeedContainer,
	}

	details := []FormulaDetail{}
	for _, detail := range payload.Detail {
		tempRec := FormulaDetail{
			FormulaRecID:       0,
			MaterialID:         detail.MaterialID,
			MaterialWeight:     detail.MaterialWeight,
			MaterialPercentage: detail.MaterialPercentage,
			Sequence:           detail.Sequence,
			AllowableError:     detail.AllowableError,
			Remark:             detail.Remark,
		}
		details = append(details, tempRec)
	}

	if err := NewFormulaRecProvider().UpdateFormula(header, details); err != nil {
		// TODO: error handling
		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_FORMULA_UPDATE, MsgBody: err.Error()}
	}

	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_FORMULA_UPDATE, MsgBody: "ok"}

}

// 获取配方
func (p getFormulaListNotifier) Handle(mgr *SrvMgr) {
	// Do something for this event
	l.Log.Debug("Handle getFormulaListNotifier called")
	recs, _ := NewFormulaRecProvider().GetFormulaDataList()

	var typesStr string
	var err error
	if typesStr, err = json.MarshalToString(recs); err != nil {
		l.Log.Error(err)

	}
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_FORMULA_LIST, MsgBody: typesStr}

}

// 传入原来的序号从payload中找出对应的ReqFormulaWgtRecDetail

func getDetailRecFromPayload(payload ReqFormulaWgtRec, seq int) ReqFormulaWgtRecDetail {
	for _, detail := range payload.RecDetail {
		if detail.Sequence == seq {
			return detail
		}
	}
	return ReqFormulaWgtRecDetail{}
}

// 新增配方称重记录
func (p addFormulaWgtRecNotifier) Handle(mgr *SrvMgr, payload ReqFormulaWgtRec) {
	// Do something for this event
	l.Log.Debug("Handle addFormulaWgtRecNotifier called")
	//先找出配方信息 写记录的时候，将原来的配方信息也写进去
	fmaId := payload.RecHeader.FormulaID
	fmaInfo, err := NewFormulaRecProvider().GetFormulaListByFormulaID(fmaId)

	if err != nil {
		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_FORMULA_REC_ADD, MsgBody: "Formula not found"}
		return
	}

	newFmaHeader := FormulaWgtRecHeader{
		RecordID:          payload.RecHeader.RecordID,
		RecordSaveTime:    time.Now(),
		Operator:          payload.RecHeader.Operator,
		FormulaID:         fmaId,
		FormulaName:       fmaInfo.Header.FormulaHeader.FormulaName,
		FormulaMode:       fmaInfo.Header.FormulaHeader.FormulaMode,
		FormulaTypeId:     fmaInfo.Header.FormulaHeader.CategoryID,
		FormulaTypeName:   fmaInfo.Header.FormulaCategoryName,
		TotalWeight:       payload.RecHeader.TotalWeight,
		ActualTotalWeight: payload.RecHeader.ActualTotalWeight,
		TotalWeightUnit:   payload.RecHeader.TotalWeightUnit,
		MaterialCount:     fmaInfo.Header.FormulaHeader.MaterialCount,
		Error:             0.0,
		IsQualified:       payload.RecHeader.IsQualified,
		ActualFmaTotalWgt: payload.RecHeader.ActualFmaTotalWgt, //实际配方总重量,包括修正后需要的重量
		IsEncrypted:       fmaInfo.Header.FormulaHeader.IsEncrypted,
		NeedContainer:     fmaInfo.Header.FormulaHeader.NeedContainer,
		FormulaCreatedAt:  fmaInfo.Header.FormulaHeader.CreatedAt,
		FormulaUpdatedAt:  fmaInfo.Header.FormulaHeader.UpdatedAt,
		FormulaCreatedBy:  fmaInfo.Header.FormulaHeader.CreatedBy,
		FormulaUpdatedBy:  fmaInfo.Header.FormulaHeader.UpdatedBy,
		FormulaRemark:     fmaInfo.Header.FormulaHeader.Remark,
		FormulaRemark1:    fmaInfo.Header.FormulaHeader.Remark1,
		ScaleId:           payload.RecHeader.ScaleId,
		ScaleName:         payload.RecHeader.ScaleName,
		ScaleModel:        payload.RecHeader.ScaleModel,
		ScaleSn:           payload.RecHeader.ScaleSn,
	}

	//插入头
	if err := NewFormulaRecProvider().InsertFormulaWgtHeader(newFmaHeader); err != nil {
		// TODO: error handling
		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_FORMULA_REC_ADD, MsgBody: err.Error()}
		return
	}
	//如果有容器
	if fmaInfo.Header.FormulaHeader.NeedContainer {
		detailRec := getDetailRecFromPayload(payload, 0)
		newFmaDetail := FormulaWgtRecDetail{
			RecordID:           payload.RecHeader.RecordID,
			MaterialID:         "-",
			MaterialName:       "-",
			MaterialTypeID:     0,
			MaterialTypeName:   "-",
			Ingredient:         "-",
			MaterialCreatedAt:  time.Now(),
			MaterialUpdatedAt:  time.Now(),
			MaterialCreatedBy:  "-",
			MaterialUpdatedBy:  "-",
			MaterialRemark:     "-",
			MaterialRemark1:    "-",
			TargetWgt:          0,
			MaterialWeight:     0,
			MaterialPercentage: 0,
			Sequence:           0,
			AllowableError:     0,
			ActualWeight:       detailRec.ActualWeight,
			ActualPercentage:   0,
			ActualErrorWgt:     0,
			ActualErrorPct:     0,
			IsQualified:        detailRec.IsQualified,
			LastWeighingTime:   time.Now(),
		}
		//插入容器
		if err := NewFormulaRecProvider().InsertFormulaWgtBody(newFmaDetail); err != nil {
			// TODO: error handling
			mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_FORMULA_REC_ADD, MsgBody: err.Error()}
			return
		}
	}

	//插入体
	for _, detail := range fmaInfo.Details {
		//先找到payload中对应的detail
		newFmaDetail := FormulaWgtRecDetail{}
		//先找出是不是容器

		//不是容器
		//计算出配方的单重
		var rawWgt float64 = 0.0
		if fmaInfo.Header.FormulaHeader.FormulaMode == "pct" {
			rawWgt = math.Round(payload.RecHeader.ActualFmaTotalWgt*detail.FormulaDetail.MaterialPercentage/100*1000) / 1000
		} else {
			rawWgt = detail.FormulaDetail.MaterialWeight
		}
		detailRec := getDetailRecFromPayload(payload, detail.FormulaDetail.Sequence)
		newFmaDetail = FormulaWgtRecDetail{
			RecordID:           payload.RecHeader.RecordID,
			MaterialID:         detailRec.MaterialID,
			MaterialName:       detail.RawMaterialTypeName.RawMaterial.MaterialName,
			MaterialTypeID:     detail.RawMaterialTypeName.RawMaterial.CategoryID,
			MaterialTypeName:   detail.RawMaterialTypeName.RawCategoryName,
			Ingredient:         detail.RawMaterialTypeName.RawMaterial.Ingredient,
			MaterialCreatedAt:  detail.RawMaterialTypeName.RawMaterial.CreatedAt,
			MaterialUpdatedAt:  detail.RawMaterialTypeName.RawMaterial.UpdatedAt,
			MaterialCreatedBy:  detail.RawMaterialTypeName.RawMaterial.CreatedBy,
			MaterialUpdatedBy:  detail.RawMaterialTypeName.RawMaterial.UpdatedBy,
			MaterialRemark:     detail.RawMaterialTypeName.RawMaterial.Remark,
			MaterialRemark1:    detail.RawMaterialTypeName.RawMaterial.Remark1,
			TargetWgt:          detailRec.TargetWgt,
			MaterialWeight:     rawWgt,
			MaterialPercentage: detail.FormulaDetail.MaterialPercentage,
			Sequence:           detail.FormulaDetail.Sequence,
			AllowableError:     detail.FormulaDetail.AllowableError,
			ActualWeight:       detailRec.ActualWeight,
			ActualPercentage:   detailRec.ActualPercentage,
			ActualErrorWgt:     detailRec.ActualErrorWgt,
			ActualErrorPct:     detailRec.ActualErrorPct,
			IsQualified:        detailRec.IsQualified,
			LastWeighingTime:   time.Now(),
		}

		//插入详细
		if err := NewFormulaRecProvider().InsertFormulaWgtBody(newFmaDetail); err != nil {
			// TODO: error handling
			mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_FORMULA_REC_ADD, MsgBody: err.Error()}
			return
		}

	}

	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_FORMULA_REC_ADD, MsgBody: "ok"}

}

// 获取配方称重记录
func (p getFormulaWgtRecListNotifier) Handle(mgr *SrvMgr) {
	// Do something for this event
	l.Log.Debug("Handle getFormulaWgtRecListNotifier called")
	recs, _ := NewFormulaRecProvider().GetFormulaWgtRecList()

	var typesStr string
	var err error
	if typesStr, err = json.MarshalToString(recs); err != nil {
		l.Log.Error(err)

	}
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_FORMULA_REC_LIST, MsgBody: typesStr}

}

// 删除配方
func (p delFormulaNotifier) Handle(mgr *SrvMgr, payload ReqDelFmaData) {
	// Do something for this event
	l.Log.Debug("Handle delFormulaNotifier called")
	rec := payload.RecId
	if err := NewFormulaRecProvider().DeleteFormula(rec); err != nil {
		// TODO: error handling
		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_FORMULA_DELETE, MsgBody: err.Error()}
	}
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_FORMULA_DELETE, MsgBody: "ok"}

}

// 增加流速
func (p addFlowRateNotifier) Handle(mgr *SrvMgr, payload ReqFlowRateRec) {
	// Do something for this event
	l.Log.Debug("Handle addFlowRateNotifier called")
	var header FlowRateHeader = FlowRateHeader{
		TotalWeight:     payload.RecHeader.TotalWeight,
		TotalTime:       payload.RecHeader.TotalTime,
		AverageFlowRate: payload.RecHeader.AverageFlowRate,
		MinFlowRate:     payload.RecHeader.MinFlowRate,
		MaxFlowRate:     payload.RecHeader.MaxFlowRate,
		WgtUnit:         payload.RecHeader.WgtUnit,
	}
	if err := NewFlowRateProvider().infoPb.AddFlowRateHeader(&header); err != nil {
		// TODO: error handling
		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_FLOW_RATE_ADD, MsgBody: err.Error()}
	}

	//找出头表中最大的recid
	recId, err := NewFlowRateProvider().infoPb.FindMaxRecId()
	if err != nil {
		// TODO: error handling
		recId = 0
	}
	for _, detail := range payload.RecDetail {
		tempRec := FlowRateDetail{
			HeaderId: recId,
			Id:       detail.Id,
			Rate:     detail.Rate,
			Time:     detail.Time,
		}
		//插入详细
		if err := NewFlowRateProvider().infoPb.AddFlowRateDetail(&tempRec); err != nil {
			// TODO: error handling
			mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_FLOW_RATE_ADD, MsgBody: err.Error()}
		}
	}

	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_FLOW_RATE_ADD, MsgBody: "ok"}

}

// 获取流速
func (p getFlowRateListNotifier) Handle(mgr *SrvMgr) {
	// Do something for this event
	l.Log.Debug("Handle getFlowRateNotifier called")
	recs, _ := NewFlowRateProvider().GetFlowRateList()

	var typesStr string
	var err error
	if typesStr, err = json.MarshalToString(recs); err != nil {
		l.Log.Error(err)

	}
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_FLOW_RATE_LIST, MsgBody: typesStr}

}
