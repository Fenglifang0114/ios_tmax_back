package svc

import (
	"fmt"
	"log"
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
	productPd   *ProductRecProvider
	userPd      *UserRecProvider
	wifiPd      *WifiRecProvider
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
				if client.scaleId != 0 {
					if h.scales[client.scaleId] != nil {
						h.scales[client.scaleId].HandleClientDisconnect()
					}

				}

				client.Close()

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
