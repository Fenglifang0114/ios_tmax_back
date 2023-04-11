package svc

import (
	"fmt"
	"time"

	"go.bug.st/serial"

	"tmaxsrv/log"
)

var nextScaleId int64 = 1 // this scale id will be incremented as new scale is added, 0 is reserved for not used
type ScaleMgr struct {
	srvMgr *SrvMgr
	scales map[int64]*Scale // map with scale id
	connPb *ScaleConnProvider
	recPb  *ScaleRecProvider
	medias []*ScaleConnMedia // scale connections meida
}

func NewScaleMgr() *ScaleMgr {
	connPb := NewScaleConnProvider()
	recPb := NewScaleRecProvider()
	scales := make(map[int64]*Scale)
	return &ScaleMgr{connPb: connPb, recPb: recPb, scales: scales}
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
}

type portListedNotifier struct {
}

type scaleListedNotifier struct {
}

type addScaleNotifier struct {
}

type modifyScaleNotifier struct {
}

type productListedNotifier struct {
}

type addProductNotifier struct {
}

type delProductNotifier struct {
}

type modifyProductNotifier struct {
}

type userListedNotifier struct {
}

type addUserNotifier struct {
}

type delUserNotifier struct {
}

type modifyUserNotifier struct {
}

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
	//scaleMedias, _ := NewScaleConnProvider().GetScaleConnsList()
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
	if err := mSrvMgr.scaleMgr.AddScale(payload); err != nil {
		log.Log.Errorf("%v\n", err)
		// TODO: return NACK to requester
	}
	// TODO: check if this connection is already existed
	scaleConns, _ := NewScaleConnProvider().GetScaleConnsList()
	var scalesStr string
	var err error
	if scalesStr, err = json.MarshalToString(scaleConns); err != nil {
		log.Log.Errorf("%v\n", err)
		// TODO: error handling
	}
	// send ports list back to requestee
	mSrvMgr.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_SCALES_LIST, MsgBody: scalesStr}
}

func (p modifyScaleNotifier) Handle(payload ReqModifyScale) {
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

// Run function will scan the scale from the scale list that from database, will inform srvMgrif any scale's online state is changed
func (s *ScaleMgr) Run() {
	// list serial from time to time to check if the port that connecting to scale is varied
	s.medias, _ = s.connPb.GetScaleConnsList() // scaleId "0000" is for get all scale connections
	for {
		// construct scale instance if it doesn't exist
		for _, conn := range s.medias {
			if conn.scale == nil {
				// new scale and assign scaleid to the instance
				var scale *Scale
				scale, _ = NewScale(s, conn, conn.ScaleModel, conn.ScaleSn, false) // TODO: check this blocks
				scale.Id = nextScaleId
				conn.scale = scale
				conn.ScaleId = scale.Id
				s.scales[scale.Id] = scale
				s.srvMgr.addScale <- scale // register new scale instance to srvMgr
				nextScaleId++
			}
		}

		ports, _ := getPortsList()
		// portsNotInUse := handlePortState(&ports, &s.conns)
		_ = handlePortState(&ports, &s.medias)
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

func contains(s []string, e string) (int, bool) {
	for i, a := range s {
		if a == e {
			return i, true
		}
	}
	return -1, false
}

func getPortsList() ([]string, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		log.Log.Error(err)
	}
	if len(ports) == 0 {
		//log.Fatal("No serial ports found!")
	}
	// for _, port := range ports {
	// 	fmt.Printf("Found port: %v\n", port)
	// }

	return ports, err
}

func (m *ScaleMgr) GetScale(id int64) (*Scale, error) {
	scale := m.scales[id]
	if scale == nil {
		return nil, fmt.Errorf("can't find scale")
	}

	return scale, nil
}

// for user to add a scale, should avoid to overwrite existing scale
func (m *ScaleMgr) AddScale(req ReqAddScale) error {
	// check if the scaleConn is existing via checking the scale's model and scale's sn

	return nil
}

// for user to delete a scale
func (m *ScaleMgr) DelScale(id string) error {
	return nil
}

// for user to update a scale
func (s *ScaleMgr) UpdateScale(req ReqModifyScale) error {
	id := req.ScaleId
	tmpScale := s.scales[id]
	if tmpScale == nil {
		return fmt.Errorf("can't find scale wit id: %v", id)
	}

	conn := tmpScale.Conn
	if conn == nil {
		return fmt.Errorf("can't find connection associated with the scale Id")
	}
	conn.MediaConf = req.MediaConf
	// TODO: change scale's mediaConf
	s.scales[id].ModifyMedia(req.MediaConf)
	client := s.srvMgr.clientScales[s.srvMgr.scales[id]]
	// remove the conn then add new one s.conns
	if client != nil && client.scaleId == id {
		s.srvMgr.unregister <- client
		if client.conn != nil {
			client.conn.Close()
		} // terminate the socket that associate with the scale
	}
	// s.srvMgr.removeScale <- s.scales[id]                                                // remove the old scale
	// scale, _ := NewScale(s.srvMgr.scaleMgr, conn, conn.ScaleModel, conn.ScaleSn, false) // TODO: check this blocks
	// scale.Id = nextScaleId
	// conn.ScaleId = scale.Id
	// s.scales[scale.Id] = scale
	// s.srvMgr.addScale <- scale // register new scale instance to srvMgr
	// nextScaleId++
	s.srvMgr.scaleMgr.ModifyMediaList(id, conn.MediaConf)
	s.connPb.connPb.UpdateScaleConn(*conn)

	return nil
}

func (s *ScaleMgr) GetScaleRecs(scale *Scale) ([]ScaleRec, error) {
	var recs []ScaleRec
	var err error
	if recs, err = s.recPb.GetRecsList(*scale); err != nil {
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
