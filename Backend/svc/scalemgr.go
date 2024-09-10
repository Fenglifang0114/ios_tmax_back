package svc

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"tmaxsrv/comm"
	"tmaxsrv/log"
)

var nextScaleId int64 = 1 // this scale id will be incremented as new scale is added, 0 is reserved for not used
type ScaleMgr struct {
	mu                sync.Mutex
	srvMgr            *SrvMgr
	scales            map[int64]*Scale // map with scale id
	connPb            *ScaleConnProvider
	recPb             *ScaleRecProvider
	infoPb            *ScaleInfosProvider
	pluFilePb         *PluRecProvider
	recCheckWeigherPb *ScaleRecCheckWeigherProvider
	recTakeInPb       *ScaleRecTakeInProvider
	recTakeOutPb      *ScaleRecTakeOutProvider
	medias            []*ScaleConnMedia // scale connections meida
	detailPb          *DetailRecProvider
}

func NewScaleMgr() *ScaleMgr {
	connPb := NewScaleConnProvider()
	recPb := NewScaleRecProvider()
	infoPb := NewScaleInfosProvider()
	pluFilePb := NewPluRecProvider()
	recCheckWeigherPb := NewScaleRecCheckWeigherProvider()
	recTakeInPb := NewScaleRecTakeInProvider()
	recTakeOutPb := NewScaleRecTakeOutProvider()
	scales := make(map[int64]*Scale)
	detailPb := NewDetailRecProvider()
	return &ScaleMgr{connPb: connPb, recPb: recPb, infoPb: infoPb, pluFilePb: pluFilePb, recCheckWeigherPb: recCheckWeigherPb, recTakeInPb: recTakeInPb, recTakeOutPb: recTakeOutPb, scales: scales, detailPb: detailPb}
}

func (s *ScaleMgr) SetSrvMsg(srvMgr *SrvMgr) {
	s.srvMgr = srvMgr
}

func init() {
	createNotifier := portListedNotifier{}
	portsListed.Register(createNotifier)

	createScaleListNotifier := scaleListedNotifier{}
	scalesListed.Register(createScaleListNotifier)

	createAddScaleNotifier := addScaleNotifier{}
	scaleAdded.Register(createAddScaleNotifier)

	createDelScaleNotifier := delScaleNotifier{}
	scaleDeleted.Register(createDelScaleNotifier)

	createModifyScaleNotifier := modifyScaleNotifier{}
	scaleModified.Register(createModifyScaleNotifier)

	createProductListNotifier := productListedNotifier{}
	productsListed.Register(createProductListNotifier)

	createAddProductNotifier := addProductNotifier{}
	productAdded.Register(createAddProductNotifier)

	createDelProductNotifier := delProductNotifier{}
	productDeleted.Register(createDelProductNotifier)

	createModifyProductNotifier := modifyProductNotifier{}
	productModified.Register(createModifyProductNotifier)

	createUserListNotifier := userListedNotifier{}
	usersListed.Register(createUserListNotifier)

	createAddUserNotifier := addUserNotifier{}
	userAdded.Register(createAddUserNotifier)

	createDelUserNotifier := delUserNotifier{}
	userDeleted.Register(createDelUserNotifier)

	createModifyUserNotifier := modifyUserNotifier{}
	userModified.Register(createModifyUserNotifier)

	createDetailListNotifier := detailListedNotifier{}
	detailListed.Register(createDetailListNotifier)
}

type portListedNotifier struct{}

type scaleListedNotifier struct{}

type addScaleNotifier struct{}

type delScaleNotifier struct{}

type modifyScaleNotifier struct{}

type productListedNotifier struct{}

type addProductNotifier struct{}

type delProductNotifier struct{}

type modifyProductNotifier struct{}

type userListedNotifier struct{}

type addUserNotifier struct{}

type delUserNotifier struct{}

type modifyUserNotifier struct{}

type detailListedNotifier struct{}

func (p portListedNotifier) Handle() {
	// Do something for this event
	log.Log.Debug("Handle portListedNotifier called")
	// Do something with this event
	ports, _ := getPortsList()
	var portsStr string
	var err error
	if portsStr, err = json.MarshalToString(ports); err != nil {
		// fmt.Printf("%v\n", err)
		// TODO: error handling
	}
	// send ports list back to requestee
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_PORTS_LIST, MsgBody: portsStr}
}

func (p scaleListedNotifier) Handle(scaleMgr *ScaleMgr) {
	// Do something for this event
	log.Log.Debug("Handle scaleListedNotifier called")
	// Do something with this event
	// scaleMedias, _ := NewScaleConnProvider().GetScaleConnsList()
	scaleMedias := scaleMgr.medias
	var scalesStr string
	var err error
	if scalesStr, err = json.MarshalToString(scaleMedias); err != nil {
		log.Log.Errorf("%v\n", err)
		// TODO: error handling
	}
	// send ports list back to requestee
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_SCALES_LIST, MsgBody: scalesStr}
}

func (p addScaleNotifier) Handle(payload ReqAddScale) {
	// Do something for this event
	log.Log.Debug("Handle addScaleNotifier called")
	var scalesStr string
	var err error

	if err := mSrvMgr.scaleMgr.AddScale(payload); err != nil {
		log.Log.Errorf("%v\n", err)
		// return NACK to requester
		scalesStr = err.Error()
		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_SCALE_ADD, MsgBody: scalesStr}
		return
	}
	// TODO: check if this connection is already existed
	scaleConns, _ := NewScaleConnProvider().GetScaleConnsList()
	if scalesStr, err = json.MarshalToString(scaleConns); err != nil {
		log.Log.Errorf("%v\n", err)
		// TODO: error handling
	}
	// send ports list back to requestee
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_SCALES_LIST, MsgBody: scalesStr}
}

func (p detailListedNotifier) Handle(scaleMgr *ScaleMgr) {
	// Do something for this event
	log.Log.Debug("Handle detailListedNotifier called")
	// Do something with this event
	detailLists, _ := NewDetailRecProvider().GetRecsList()

	var detailsStr string
	var err error
	if detailsStr, err = json.MarshalToString(detailLists); err != nil {
		log.Log.Errorf("%v\n", err)
		// TODO: error handling
	}
	// send ports list back to requestee
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_DETAIL_LIST, MsgBody: detailsStr}
}

func (p delScaleNotifier) Handle(payload ReqDelScale) { //修改秤的属性
	// Do something for this event
	log.Log.Debug("Handle modifyScaleNotifier called")
	if err := mSrvMgr.scaleMgr.DelScale(payload.ScaleId); err != nil {
		log.Log.Errorf("%v\n", err)
		resp := MgrRespMsg{IsAck: true, AckData: err.Error()}
		jsonStr, _ := json.MarshalToString(resp)
		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_SCALE_DEL, MsgBody: jsonStr}
		return
	}
	resp := MgrRespMsg{IsAck: true, AckData: "ok"}
	jsonStr, _ := json.MarshalToString(resp)
	// send ports list back to requestee
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_SCALE_DEL, MsgBody: jsonStr}
}

func (p modifyScaleNotifier) Handle(payload ReqModifyScale) { //修改秤的属性
	// Do something for this event
	log.Log.Debug("Handle modifyScaleNotifier called")
	if err := mSrvMgr.scaleMgr.UpdateScale(payload); err != nil {
		log.Log.Errorf("%v\n", err)
		resp := MgrRespMsg{IsAck: true, AckData: err.Error()}
		jsonStr, _ := json.MarshalToString(resp)
		mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_SCALE_MODIFY, MsgBody: jsonStr}
	}
	resp := MgrRespMsg{IsAck: true, AckData: ""}
	jsonStr, _ := json.MarshalToString(resp)
	// send ports list back to requestee
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_SCALE_MODIFY, MsgBody: jsonStr}
}

// Run function will scan the scale from the scale list that from database, will inform srvMgr if any scale's online state is changed
func (s *ScaleMgr) Run() {
	// list serial from time to time to check if the port that connecting to scale is varied
	s.medias, _ = s.connPb.GetScaleConnsList() // scaleId "0000" is for get all scale connections
	for {

		// construct scale instance if it doesn't exist
		for _, conn := range s.medias {
			if conn.scale == nil {
				// new scale and assign scaleid to the instance
				var scale *Scale
				scale, _ = NewScale(s, conn, conn.ScaleCat, conn.ScaleModel, conn.ScaleSn, false)
				// scale.Id = nextScaleId
				scale.Id = conn.ScaleId
				conn.scale = scale
				conn.ScaleId = scale.Id
				s.scales[scale.Id] = scale
				s.srvMgr.addScale <- scale // register new scale instance to srvMgr

				if conn.ScaleId >= nextScaleId {
					nextScaleId = conn.ScaleId
					nextScaleId++
				}

			}
		}

		ports, _ := getPortsList()
		// portsNotInUse := handlePortState(&ports, &s.conns)
		_ = handlePortState(&ports, &s.medias)
		// _ = handleNetState(&s.medias)
		// fmt.Printf("%v\n", portsNotInUse)
		// TODO: scan ports for finding scales and handle new finding scales

		// TODO: handle scales that offline
		time.Sleep(2 * time.Second) // detect connectivity every 2 seconds
	}
}

func (s *ScaleMgr) ModifyMediaList(scaleId int64, conf MediaConf) error {
	isFound := false
	for i, media := range s.medias {
		if media.ScaleId == scaleId {
			s.medias[i].MediaConf = conf
			isFound = true
			break
		}
	}
	if !isFound {
		return fmt.Errorf("Not found the scale conf")
	}

	return nil
}

func (s *ScaleMgr) ModifyScaleInfo(scaleId int64, modelName string, sn string) error {
	isFound := false
	for i, media := range s.medias {
		if media.ScaleId == scaleId {
			s.medias[i].ScaleModel = modelName
			s.medias[i].ScaleSn = sn
			isFound = true
			break
		}
	}
	if !isFound {
		return fmt.Errorf("Not found the scale conf")
	}

	return nil
}

func (s *ScaleMgr) AddMediaList(scaleId int64, conn ScaleConnMedia) error {
	isFound := false
	for i, media := range s.medias {
		if media.ScaleId == scaleId {
			s.medias[i].MediaConf = conn.MediaConf
			isFound = true
			break
		}
	}
	if isFound {
		return fmt.Errorf("exsit the scale conf")
	}
	s.medias = append(s.medias, &conn)

	return nil
}

func (s *ScaleMgr) DelMediaList(scaleId int64, conn ScaleConnMedia) error {
	result := []*ScaleConnMedia{}
	for _, m := range s.medias {
		if m.ScaleId != scaleId && m.MediaConf != conn.MediaConf {
			result = append(result, m)
		}
	}
	s.medias = result
	return nil

}

func sContainsConnMedia(conn *ScaleConnMedia) bool {
	return conn.scale != nil
}

func handlePortState(inPorts *[]string, conns *[]*ScaleConnMedia) (portsNotInUse []string) {
	var comInfo ComInfo

	for _, port := range *inPorts {
		isFound := false
		for _, conn := range *conns {
			if conn.TMedia == MEDIA_COM {
				if err := json.UnmarshalFromString(conn.MediaConf.MediaInfoJson, &comInfo); err != nil {
					return nil // TODO: check error
				}
				if port == comInfo.DevPath {
					isFound = true
					break
				}
			}
		}
		if !isFound {
			portsNotInUse = append(portsNotInUse, port)
		}
	}

	for i, conn := range *conns {
		if conn.TMedia == MEDIA_COM {
			if err := json.UnmarshalFromString(conn.MediaConf.MediaInfoJson, &comInfo); err != nil {
				return nil // TODO: check error
			}
			if _, isFound := contains(*inPorts, comInfo.DevPath); isFound {
				continue
			}
			(*conns)[i].IsOnline = false
		}

	}
	return
}

func handleNetState(conns *[]*ScaleConnMedia) (netsNotInUse []string) {

	var netInfo NetInfo

	for i, conn := range *conns {
		if conn.TMedia == MEDIA_NET {
			if err := json.UnmarshalFromString(conn.MediaConf.MediaInfoJson, &netInfo); err != nil {
				return nil // TODO: check error
			}
			if i >= len(*conns) {
				return nil
			}
			if conn.scale == nil {
				return nil
			}

			if conn.scale.MyNet == nil {
				return nil
			}

			if conn.scale.MyNet.conn != nil && conn.scale.MyNet.isAlive {
				fmt.Printf("already connect :", conn.scale.MyNet.ip)
				continue
			}

			var err error
			if !conn.scale.MyNet.toQuit {
				if conn.scale.MyNet != nil {
					conn.scale.MyNet.conn, err = conn.scale.MyNet.reconnect()
					fmt.Printf("reconnect ip :", conn.scale.MyNet.ip)
					if err == nil {
						conn.scale.MyNet.isAlive = true
						continue
					}
					if i >= len(*conns) {
						return nil
					}

					(*conns)[i].IsOnline = false
					conn.scale.MyNet.isAlive = false
				}

			}

		}

	}

	return
}

func contains(s []string, e string) (int, bool) {
	for i, a := range s {
		if a == e {
			return i, true
		}
	}
	return -1, false
}

func (m *ScaleMgr) GetScale(id int64) (*Scale, error) {
	scale := m.scales[id]
	if scale == nil {
		return nil, fmt.Errorf("can't find scale")
	}

	return scale, nil
}

// 随机写个sn初始化的时候，并不知道SN是什么
func getSn() string {
	timestamp := time.Now().Unix()
	// 将时间戳转换为字符串
	strTimestamp := strconv.FormatInt(timestamp, 10)
	return strTimestamp

}

// for user to add a scale, should avoid to overwrite existing scale
func (s *ScaleMgr) AddScale(req ReqAddScale) error {
	// check if the scaleConn is existing via checking the scale's model and scale's sn
	//TODO:要用于增加管理的秤  202406
	s.medias, _ = s.connPb.GetScaleConnsList()
	var netInfo NetInfo
	var reqNetInfo NetInfo
	//现在只新增网络秤
	if req.MediaConf.Type != MEDIA_NET {
		return fmt.Errorf("only can add net scale")
	}
	if err := json.UnmarshalFromString(req.MediaConf.MediaInfoJson, &reqNetInfo); err != nil {
		return err
	}
	for _, conn := range s.medias {
		if conn.MediaConf.Type == MEDIA_NET {
			if err := json.UnmarshalFromString(conn.MediaConf.MediaInfoJson, &netInfo); err != nil {
				return err
			}
			if reqNetInfo.Ip == netInfo.Ip {
				return fmt.Errorf("This IP address already exists")
			}
		}
	}

	conn := &ScaleConnMedia{ScaleModel: req.ScaleModel, ScaleSn: getSn(), TMedia: req.MediaConf.Type, MediaConf: req.MediaConf}
	conn.ScaleModel = "TMax"
	conn.ScaleId = nextScaleId
	conn.IsDefault = true
	conn.IsOnline = true
	conn.ScaleCat = comm.SCALE_TMAX
	conn.ScaleName = "Scale" + strconv.FormatInt(conn.ScaleId, 10)

	var scale *Scale

	scale, _ = NewScale(s, conn, conn.ScaleCat, conn.ScaleModel, conn.ScaleSn, false)
	scale.Id = conn.ScaleId
	conn.scale = scale
	s.scales[scale.Id] = scale
	s.srvMgr.addScale <- scale // register new scale instance to srvMgr
	nextScaleId++
	s.connPb.connPb.InsertScaleConn(*conn)

	return nil

}

// for user to delete a scale
func (s *ScaleMgr) DelScale(id int64) error {
	//用于删除管理的秤  202406
	scale := s.scales[id]
	if scale == nil {
		return fmt.Errorf("can't find scale with id: %v", id)
	}
	if scale.Conn.TMedia == MEDIA_COM {
		return fmt.Errorf("serial connect can not delete")
	}

	conn := scale.Conn
	if conn == nil {
		return fmt.Errorf("can't find connection associated with the scale Id")
	}

	client := s.srvMgr.clientOfScales[scale]
	if client != nil && client.scaleId == id {
		s.srvMgr.unregister <- client
		if client.conn != nil {
			client.conn.Close()
		} // terminate the socket that associate with the scale
	}
	// remove the conn then add new one s.conns

	scale.Close()
	s.srvMgr.removeScale <- scale
	s.connPb.connPb.DeleteScaleConn(*conn)
	if s.scales[scale.Id] != nil { // scale not existing
		s.scales[scale.Id] = nil
	}

	return nil

}

// for user to update a scale
func (s *ScaleMgr) UpdateScale(req ReqModifyScale) error {
	id := req.ScaleId
	scale := s.scales[id]
	if scale == nil {
		return fmt.Errorf("can't find scale with id: %v", id)
	}

	if scale.Model != req.ScaleModel {
		scale.Model = req.ScaleModel
		if strings.Contains(strings.ToLower(scale.Model), "tmax") {
			scale.ScaleCat = comm.SCALE_TMAX
		} else {
			scale.ScaleCat = comm.SCALE_T2200
		}
		composer := cmdComposerFuncMap[scale.ScaleCat]
		scale.composer = &composer
	}
	conn := scale.Conn
	if conn == nil {
		return fmt.Errorf("can't find connection associated with the scale Id")
	}
	conn.MediaConf = req.MediaConf
	// TODO: change scale's mediaConf
	s.scales[id].ModifyMedia(req.MediaConf)
	// client := s.srvMgr.clientOfScales[s.srvMgr.scales[id]]
	// // remove the conn then add new one s.conns
	// if client != nil && client.scaleId == id {
	// 	s.srvMgr.unregister <- client
	// 	if client.conn != nil {
	// 		client.conn.Close()
	// 	} // terminate the socket that associate with the scale
	// }
	//---------------------上面的先删除试试
	// s.srvMgr.removeScale <- s.scales[id]                                                // remove the old scale
	// scale, _ := NewScale(s.srvMgr.scaleMgr, conn, conn.ScaleModel, conn.ScaleSn, false) // TODO: check this blocks
	// scale.Id = nextScaleId
	// conn.ScaleId = scale.Id
	// s.scales[scale.Id] = scale
	// s.srvMgr.addScale <- scale // register new scale instance to srvMgr
	// nextScaleId++
	s.srvMgr.scaleMgr.ModifyMediaList(id, conn.MediaConf)
	conn.ScaleCat = scale.ScaleCat
	conn.ScaleModel = scale.Model
	s.connPb.connPb.UpdateScaleConn(*conn)
	return nil
}

// for user to update a scale
func (s *ScaleMgr) UpdateScaleSn(req ReqModifyScaleSn) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := req.ScaleId
	scale := s.scales[id]
	if scale == nil {
		return fmt.Errorf("can't find scale with id: %v", id)
	}

	if scale.Model == req.ScaleModel && scale.Sn == req.Sn {
		return nil
	}
	scale.Model = req.ScaleModel
	scale.Sn = req.Sn
	conn := scale.Conn
	conn.ScaleModel = scale.Model
	conn.ScaleSn = scale.Sn
	s.srvMgr.scaleMgr.ModifyScaleInfo(id, req.ScaleModel, req.Sn)
	s.connPb.connPb.UpdateScaleInfo(*conn)
	return nil
}

func (s *ScaleMgr) GetScaleRecs(scale *Scale) ([]ScaleRec, error) {
	var recs []ScaleRec
	var err error
	if recs, err = s.recPb.GetRecsList(*scale, "", "", ""); err != nil {
		return recs, err
	}
	return recs, nil
}

func (s *ScaleMgr) InsertScaleRec(rec ScaleRec) error {
	return s.recPb.InsertRec(rec)
}

func (s *ScaleMgr) DeleteScaleRec(recId uint) error {
	return s.recPb.DeleteRec(recId)
}
