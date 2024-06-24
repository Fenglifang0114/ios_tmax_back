package svc

import (
	"math/big"
	"strings"

	jsoniter "github.com/json-iterator/go"

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
	recvScaleMgrMsg chan *ScaleMgrRespMsg
	// Inbound messages from the scale's notification
	recvScaleNotifyMsg chan ScaleRespMsg
	// ScaleId to Scale map, used to access the scale
	scales map[int64]*Scale
	// scale to client map, used to send message from scale to the associated client
	clientOfScales map[*Scale]*Client
	// quitch channel to close this application
	quitch chan bool
	// productPb
	productPd   *ProductRecProvider
	userPd      *UserRecProvider
	uiConfig    *UiConfig
	modeSetting *ModeSettingProvider
}

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

	var licKey string
	var err error
	var licKeyList []string

	if licKey, err = lic.ReadLicFile(comm.LICENSE_FILE); err != nil || len(licKey) < 74 {
		var temp []string
		licKeyList = temp
		gIsKeyValid, gMachineId, gLicValidDate, gModuleName = lic.IsKeyValid("d7a0a41239d92ee1724cd1a311ffffff2023-05-2594df26ebd828dbff03ede5f76effffff")
		gLicenseInfoList = append(gLicenseInfoList, LicenseInfo{Id: gMachineId, ValidDate: gLicValidDate, ModuleName: gModuleName, IsValid: gIsKeyValid})
	} else {
		licKeyList = strings.Split(licKey, "\r\n")
		for _, item := range licKeyList {
			if len(item) == 74 || len(item) == 78 {
				gIsKeyValid, gMachineId, gLicValidDate, gModuleName = lic.IsKeyValid(item)
				if gIsKeyValid {
					gLicenseInfoList = append(gLicenseInfoList, LicenseInfo{Id: gMachineId, ValidDate: gLicValidDate, ModuleName: gModuleName, IsValid: gIsKeyValid})
				}

			}

		}
	}

	// if licKey, err = lic.ReadLicFile(comm.LICENSE_FILE); err != nil || len(licKey) != 74 {
	// 	l.Log.Errorf("readLicFile: %v, err: %v", comm.LICENSE_FILE, err)
	// 	gIsKeyValid, gMachineId, gLicValidDate,gModuleName = lic.IsKeyValid("d7a0a41239d92ee1724cd1a311ffffff2023-05-2594df26ebd828dbff03ede5f76effffff")
	// } else {
	// 	gIsKeyValid, gMachineId, gLicValidDate,gModuleName = lic.IsKeyValid(licKey)
	// }

	return &SrvMgr{
		scaleMgr:           scaleMgr,
		scales:             map[int64]*Scale{},
		clientOfScales:     map[*Scale]*Client{},
		recvWsClientMsg:    make(chan []byte),
		register:           make(chan *Client),
		unregister:         make(chan *Client),
		addScale:           make(chan *Scale, 2),
		removeScale:        make(chan *Scale, 2),
		recvScaleMsg:       make(chan *ScaleRespMsg, 2),
		recvScaleMgrMsg:    make(chan *ScaleMgrRespMsg, 2),
		recvScaleNotifyMsg: make(chan ScaleRespMsg, 2),
		clients:            make(map[*Client]bool),
		quitch:             quitch,
		productPd:          productPb,
		userPd:             userPb,
		uiConfig:           NewUiConfig(),
		modeSetting:        modeSettingPb,
	}
}

func (h *SrvMgr) Run() {
	for {
		select {
		case client := <-h.register:
			scaleId := client.scaleId
			if scaleId > 0 { // scaleId 0 is for management
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
				h.clientOfScales[h.scales[scaleId]] = client
				if scaleId != 0 { // id 0 is reserved for common information channel
					h.scales[scaleId].SetClient(client)
				}
				//}
			} else {
				// close(client.send) // either other client is already registered the scale or the scale is not yet registered
			}
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				// TODO: handle client disconnect
				if client.scaleId != 0 {
					h.scales[client.scaleId].HandleClientDisconnect()
				}
				client.Close()

				delete(h.clients, client)
				delete(h.clientOfScales, h.scales[client.scaleId])
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
			outData, _ := json.Marshal(scaleMgrMessage)
			if client != nil {
				client.sendCh <- outData
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
	case REQ_GET_PRODUCT_LIST:
		productsListed.Trigger(scaleMgr.srvMgr)
	case REQ_ADD_PRODUCT:
		jsonStr := req.ReqData
		var data ReqAddProduct
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
	case REQ_MODIFY_PRODUCT:
		jsonStr := req.ReqData
		var data ReqModifyProduct
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
		if err := lic.SaveKey(comm.LICENSE_FILE, req.ReqData); err != nil {
			mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_UPDATE_LICENSE, MsgBody: "fail"}
		}
		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_UPDATE_LICENSE, MsgBody: "ok"}

		// below commented: due to UI maintains records itself
		// 	case SREQ_GET_RECS: // TODO: should we check the input parameters?
		// 		var recs []ScaleRec
		// 		var err error
		// 		if recs, err = srvMgr.scaleMgr.GetScaleRecs(scale); err != nil {
		// 			srvMgr.recvScaleMsg <- &ScaleRespMsg{ScaleId: scaleId, MsgType: GET_RECS_RESP, MsgBody: err.Error()}
		// 		}
		// 		msg, _ := json.MarshalToString(recs)
		// 		srvMgr.recvScaleMsg <- &ScaleRespMsg{ScaleId: scaleId, MsgType: GET_RECS_RESP, MsgBody: msg}
		// 	case SREQ_ADD_REC:
		// 		var rec ReqAddScaleRec
		// 		if err := json.UnmarshalFromString(req.ReqData, &rec); err != nil {
		// 			srvMgr.recvScaleMsg <- &ScaleRespMsg{ScaleId: scaleId, MsgType: ADD_REC_RESP, MsgBody: err.Error()}
		// 			break
		// 		}
		// 		var scaleRec ScaleRec = ScaleRec{ScaleModel: scale.Model, ScaleSn: scale.Sn, Product: rec.Product, Weight: rec.Weight, Price: rec.Price}
		// 		if err := srvMgr.scaleMgr.InsertScaleRec(scaleRec); err != nil {
		// 			srvMgr.recvScaleMsg <- &ScaleRespMsg{ScaleId: scaleId, MsgType: ADD_REC_RESP, MsgBody: err.Error()}
		// 		}
		// 		srvMgr.recvScaleMsg <- &ScaleRespMsg{ScaleId: scaleId, MsgType: ADD_REC_RESP, MsgBody: "ok"}
		// 	case SREQ_DEL_REC:
		// 		var rec ReqDelScaleRec
		// 		if err := json.UnmarshalFromString(req.ReqData, &rec); err != nil {
		// 			srvMgr.recvScaleMsg <- &ScaleRespMsg{ScaleId: scaleId, MsgType: DEL_REC_RESP, MsgBody: err.Error()}
		// 			break
		// 		}
		// 		if err := srvMgr.scaleMgr.DeleteScaleRec(rec.RecId); err != nil {
		// 			srvMgr.recvScaleMsg <- &ScaleRespMsg{ScaleId: scaleId, MsgType: DEL_REC_RESP, MsgBody: err.Error()}
		// 			break
		// 		}
		// 		srvMgr.recvScaleMsg <- &ScaleRespMsg{ScaleId: scaleId, MsgType: DEL_REC_RESP, MsgBody: "ok"}
	}
}

// below commented: due to UI maintains records itself
// func parseToScaleReq(reqStr string) (SRequest, error) {
// 	var req SRequest
// 	if err := json.UnmarshalFromString(reqStr, &req); err != nil {
// 		log.Log.Error(err)
// 		return SRequest{}, err
// 	}

// 	return req, nil
// }

// func procToScaleReq(req SRequest, scaleId int64, srvMgr *SrvMgr, scale *Scale) error {
// 	var err error
// 	switch req.Req {
// 	case SREQ_GET_WEIGHT:
// 		// if scale.isBusy {
// 		// 	srvMgr.recvScaleMsg <- &ScaleRespMsg{ScaleId: scaleId, MsgType: WEIGHT_DATA_RESP, MsgBody: fmt.Errorf("scale is busy")}
// 		// 	return nil
// 		// }
// 		ok := scale.ReadWeight()
// 		if !ok {
// 			err = fmt.Errorf("Get weight error")
// 			msg := ScaleRespMsg{ScaleId: scaleId, MsgType: WEIGHT_DATA_RESP, MsgBody: err.Error()}
// 			msgStr, _ := json.MarshalToString(msg)
// 			scale.client.sendCh <- []byte(msgStr)
// 			// srvMgr.recvScaleMsg <- &ScaleRespMsg{ScaleId: scaleId, MsgType: WEIGHT_DATA_RESP, MsgBody: err.Error()}
// 			return nil
// 		}
// 		return nil
// 	case SREQ_ZERO:
// 		if ok := scale.PerfZero(); !ok {
// 			err = fmt.Errorf("Perform zero error")
// 			srvMgr.recvScaleMsg <- &ScaleRespMsg{ScaleId: scaleId, MsgType: ZERO_CMD_RESP, MsgBody: err.Error()}
// 		}
// 		srvMgr.recvScaleMsg <- &ScaleRespMsg{ScaleId: scaleId, MsgType: ZERO_CMD_RESP, MsgBody: ""}
// 	case SREQ_TARE:
// 		if ok := scale.PerfTare(); !ok {
// 			err = fmt.Errorf("Perform tare error")
// 			srvMgr.recvScaleMsg <- &ScaleRespMsg{ScaleId: scaleId, MsgType: TARE_CMD_RESP, MsgBody: ""}
// 		}
// 		srvMgr.recvScaleMsg <- &ScaleRespMsg{ScaleId: scaleId, MsgType: TARE_CMD_RESP, MsgBody: ""}
// 	case SREQ_REG_WEIGHT_DATA:
// 		_ = scale.RegWeightData(srvMgr.recvScaleNotifyMsg)
// 		srvMgr.recvScaleMsg <- &ScaleRespMsg{ScaleId: scaleId, MsgType: REG_WEIGHT_RESP, MsgBody: ""}
// 	case SREQ_UNREG_WEIGHT_DATA:
// 		_ = scale.UnRegWeightData(srvMgr.recvScaleNotifyMsg)
// 		srvMgr.recvScaleMsg <- &ScaleRespMsg{ScaleId: scaleId, MsgType: UNREG_WEIGHT_RESP, MsgBody: ""}
// 	case SREQ_GET_RECS: // TODO: should we check the input parameters?
// 		var recs []ScaleRec
// 		var err error
// 		if recs, err = srvMgr.scaleMgr.GetScaleRecs(scale); err != nil {
// 			srvMgr.recvScaleMsg <- &ScaleRespMsg{ScaleId: scaleId, MsgType: GET_RECS_RESP, MsgBody: err.Error()}
// 		}
// 		msg, _ := json.MarshalToString(recs)
// 		srvMgr.recvScaleMsg <- &ScaleRespMsg{ScaleId: scaleId, MsgType: GET_RECS_RESP, MsgBody: msg}
// 	case SREQ_ADD_REC:
// 		var rec ReqAddScaleRec
// 		if err := json.UnmarshalFromString(req.ReqData, &rec); err != nil {
// 			srvMgr.recvScaleMsg <- &ScaleRespMsg{ScaleId: scaleId, MsgType: ADD_REC_RESP, MsgBody: err.Error()}
// 			break
// 		}
// 		var scaleRec ScaleRec = ScaleRec{ScaleModel: scale.Model, ScaleSn: scale.Sn, Product: rec.Product, Weight: rec.Weight, Price: rec.Price}
// 		if err := srvMgr.scaleMgr.InsertScaleRec(scaleRec); err != nil {
// 			srvMgr.recvScaleMsg <- &ScaleRespMsg{ScaleId: scaleId, MsgType: ADD_REC_RESP, MsgBody: err.Error()}
// 		}
// 		srvMgr.recvScaleMsg <- &ScaleRespMsg{ScaleId: scaleId, MsgType: ADD_REC_RESP, MsgBody: "ok"}
// 	case SREQ_DEL_REC:
// 		var rec ReqDelScaleRec
// 		if err := json.UnmarshalFromString(req.ReqData, &rec); err != nil {
// 			srvMgr.recvScaleMsg <- &ScaleRespMsg{ScaleId: scaleId, MsgType: DEL_REC_RESP, MsgBody: err.Error()}
// 			break
// 		}
// 		if err := srvMgr.scaleMgr.DeleteScaleRec(rec.RecId); err != nil {
// 			srvMgr.recvScaleMsg <- &ScaleRespMsg{ScaleId: scaleId, MsgType: DEL_REC_RESP, MsgBody: err.Error()}
// 			break
// 		}
// 		srvMgr.recvScaleMsg <- &ScaleRespMsg{ScaleId: scaleId, MsgType: DEL_REC_RESP, MsgBody: "ok"}
// 	default:
// 		return fmt.Errorf("unsuported request type")
// 	}

// 	return fmt.Errorf("not processed")
// }

func (p productListedNotifier) Handle(mgr *SrvMgr) {
	// Do something for this event
	l.Log.Debug("Handle productListedNotifier called")
	// Do something with this event
	products, _ := mgr.productPd.GetRecsList()
	var productsStr string
	var err error
	if productsStr, err = json.MarshalToString(products); err != nil {
		l.Log.Error(err)
		// TODO: error handling
	}
	// send products list back to requestee
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_PRODUCTS_LIST, MsgBody: productsStr}
}

func (p addProductNotifier) Handle(mgr *SrvMgr, payload ReqAddProduct) {
	// Do something for this event
	l.Log.Debug("Handle addProductNotifier called")
	rec := ProductRec{Id: payload.Id, Product: payload.Product, WithPretare: payload.WithPretare, Pretare: payload.Pretare, Remarks: payload.Remarks}
	if err := mgr.productPd.InsertRec(rec); err != nil {
		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_PRODUCT_ADD, MsgBody: err.Error()}
	}
	// send result back to requestee
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_PRODUCT_ADD, MsgBody: ""}
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

func (p modifyProductNotifier) Handle(mgr *SrvMgr, payload ReqModifyProduct) {
	rec := ProductRec{RecId: uint(payload.RecId), Id: payload.Id, Product: payload.Product, WithPretare: payload.WithPretare, Pretare: payload.Pretare, Remarks: payload.Remarks}
	if err := NewProductRecProvider().ModifyRec(rec); err != nil {
		// TODO: error handling
	}
	resp := MgrRespMsg{IsAck: true, AckData: ""}
	jsonStr, _ := json.MarshalToString(resp)
	// send ports list back to requestee
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_SCALE_MODIFY, MsgBody: jsonStr}
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
