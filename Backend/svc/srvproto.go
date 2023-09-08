package svc

import (
	"net"

	"go.bug.st/serial"

	"tmaxsrv/comm"
)

// ********** Request for scale manager **********
type Request struct {
	Req     ReqType
	ReqData string // should be json encoded string of one of the following structures, i.e. ReqAddScale, ReqDelScale, ReqModifyScale
}

type ReqType string

const (
	REQ_GET_PORT_LIST    ReqType = "get_port_list"    // without parameter
	REQ_GET_SCALE_LIST   ReqType = "get_scale_list"   // without parameter
	REQ_GET_PRODUCT_LIST ReqType = "get_product_list" // without parameter
	REQ_GET_USER_LIST    ReqType = "get_user_list"    // without parameter

	REQ_ADD_SCALE    ReqType = "add_scale"    // with ReqAddScale parameter
	REQ_DEL_SCALE    ReqType = "del_scale"    // with ReqDelScale parameter
	REQ_MODIFY_SCALE ReqType = "modify_scale" // with ReqModifyScale parameter

	REQ_ADD_PRODUCT    ReqType = "add_product"    // with ReqAddScale parameter
	REQ_DEL_PRODUCT    ReqType = "del_product"    // with ReqDelScale parameter
	REQ_MODIFY_PRODUCT ReqType = "modify_product" // with ReqModifyScale parameter

	REQ_ADD_USER    ReqType = "add_user"    // with ReqAddScale parameter
	REQ_DEL_USER    ReqType = "del_user"    // with ReqDelScale parameter
	REQ_MODIFY_USER ReqType = "modify_user" // with ReqModifyScale parameter

	REQ_QUIT_APPLICATION ReqType = "quit_application" // without parameter
	REQ_GET_UI_CONF      ReqType = "get_ui_conf"      // without parameter
	REQ_UPDATE_UI_CONF   ReqType = "update_ui_conf"   // without parameter
	REQ_CHECK_LICENSE    ReqType = "check_license"    // without parameter
)

type ReqAddScale struct {
	ScaleModel string
	ScaleSn    string
	MediaType  MediaType
	MediaInfo  string // will be ComInfo/NetInfo/BtInfo according to the media type
}

type ReqDelScale struct {
	ScaleId int64
}

type ReqModifyScale struct {
	ScaleId int64
	// MediaType MediaType
	MediaConf  MediaConf
	ScaleModel string
}

type ReqAddProduct struct {
	Id          string
	Product     string
	WithPretare bool
	Pretare     string
	Remarks     string
}

type ReqDelProduct struct {
	RecId int64
}

type ReqModifyProduct struct {
	RecId       int64
	Id          string
	Product     string
	WithPretare bool
	Pretare     string
	Remarks     string
}

type ReqAddUser struct {
	Id       string
	Name     string
	IsFemale bool
	Phone    string
	Remarks  string
}

type ReqDelUser struct {
	RecId int64
}

type ReqModifyUser struct {
	RecId    int64
	Id       string
	Name     string
	IsFemale bool
	Phone    string
	Remarks  string
}

// ********** Response of scale manager **********
type ScaleMgrRespMsg struct {
	MsgType ScaleMgrRespMsgType
	MsgBody interface{} // MsgBody [T PortsListMsg|ScalesListMsg|MgrRespMsg|string] []T
}

type ScaleMgrRespMsgType string

const (
	SCALE_MGR_RESP_PORTS_LIST       ScaleMgrRespMsgType = "resp_ports_list"       // with response of PortsListMsg
	SCALE_MGR_RESP_SCALES_LIST      ScaleMgrRespMsgType = "resp_scales_list"      // with response of ScalesListMsg
	SCALE_MGR_RESP_SCALE_ADD        ScaleMgrRespMsgType = "resp_scale_add"        // with response of MgrRespMsg to indicate that status coreponding request procsssed
	SCALE_MGR_RESP_SCALE_DEL        ScaleMgrRespMsgType = "resp_scale_del"        // same as SCALE_MGR_RESP_SCALE_Add
	SCALE_MGR_RESP_SCALE_MODIFY     ScaleMgrRespMsgType = "resp_scale_modify"     // same as SCALE_MGR_RESP_SCALE_Add
	SCALE_MGR_RESP_PRODUCTS_LIST    ScaleMgrRespMsgType = "resp_product_list"     // with response of ScalesListMsg
	SCALE_MGR_RESP_PRODUCT_ADD      ScaleMgrRespMsgType = "resp_product_add"      // with response of MgrRespMsg to indicate that status coreponding request procsssed
	SCALE_MGR_RESP_PRODUCT_DEL      ScaleMgrRespMsgType = "resp_product_del"      // same as SCALE_MGR_RESP_SCALE_Add
	SCALE_MGR_RESP_PRODUCT_MODIFY   ScaleMgrRespMsgType = "resp_product_modify"   // same as SCALE_MGR_RESP_SCALE_Add
	SCALE_MGR_RESP_USERS_LIST       ScaleMgrRespMsgType = "resp_user_list"        // with response of ScalesListMsg
	SCALE_MGR_RESP_USER_ADD         ScaleMgrRespMsgType = "resp_user_add"         // with response of MgrRespMsg to indicate that status coreponding request procsssed
	SCALE_MGR_RESP_USER_DEL         ScaleMgrRespMsgType = "resp_user_del"         // same as SCALE_MGR_RESP_SCALE_Add
	SCALE_MGR_RESP_USER_MODIFY      ScaleMgrRespMsgType = "resp_user_modify"      // same as SCALE_MGR_RESP_SCALE_Add
	SCALE_MGR_RESP_QUIT_APPLICATION ScaleMgrRespMsgType = "resp_quit_application" // without data
	SCALE_MGR_RESP_GET_UI_CONFIG    ScaleMgrRespMsgType = "resp_get_ui_config"    // with response of UI configuration
	SCALE_MGR_RESP_UPDATE_UI_CONFIG ScaleMgrRespMsgType = "resp_update_ui_config" // without parameter
	SCALE_MGR_RESP_CHECK_LICENSE    ScaleMgrRespMsgType = "resp_check_license"    // with response of true or false
)

type PortsListMsg struct {
	PortsList []serial.Port // FIXME:
}

type ScalesListMsg struct {
	ScaleList []ScaleConnMedia
}

type ProductsListMsg struct {
	ProductList []ProductRec
}

type UsersListMsg struct {
	UserList []UserRec
}

type ScaleConnMedia struct { // connection information will be stored in database
	IsOnline   bool
	ScaleModel string
	ScaleCat   comm.ScaleCat
	ScaleSn    string
	ScaleId    int64
	TMedia     MediaType
	MediaConf  MediaConf `gorm:"embedded;embeddedPrefix:mediainfo_"`
	scale      *Scale    `gorm:"-"` // should not be stored in database
}
type MediaConf struct {
	Type          MediaType
	MediaInfoJson string // will be unmarshaled json of ComInfo, NetInfo and BtInfo
}
type MediaType int

const (
	MEDIA_COM MediaType = iota
	MEDIA_NET
	MEDIA_BT
)

type ComInfo struct {
	DevPath  string // device path of Com port, e.g. COM3
	Baud     int    // e.x. 9600
	DataBits int    // value: 7,8,9
	StopBits int    // 0: 1 stop bit, 1: 1.5 stop bits, 2: 2 stop bits
	Parity   int    // 0: no parity, 1: odd, 2: even
}
type NetInfo struct {
	Ip   string // format should be xxx.xxx.xxx.xxx
	Port int    // value should be 1-65535
}
type BtInfo struct {
	Mac  net.HardwareAddr
	Name string
}
type MgrRespMsg struct {
	IsAck   bool
	AckData string
}

type SRequest struct { // request for a scale or scale manager for records
	Req     SReqType
	ReqData string // should be json encoded string of one of the following structures, i.e. ReqScaleRec, ReqAddScaleRec, ReqDelScaleRec
}

type SReqType string

const (
	SREQ_ZERO              SReqType = "zero"
	SREQ_TARE              SReqType = "tare"
	SREQ_GET_WEIGHT        SReqType = "get_weight"
	SREQ_SEND_WT_CONT      SReqType = "send_wt_cont"
	SREQ_STOP_SEND_WT      SReqType = "stop_send_wt"
	SREQ_REG_WEIGHT_DATA   SReqType = "reg_weight_data"
	SREQ_UNREG_WEIGHT_DATA SReqType = "unreg_weight_data"
	SREQ_GET_RECS          SReqType = "get_recs" // with parameter ReqScaleRec
	SREQ_ADD_REC           SReqType = "add_rec"  // with parameter ReqAddScaleRec
	SREQ_DEL_REC           SReqType = "del_rec"  // with parameter ReqDelScaleRec
	//	SREQ_DOWN_PRN_FMT      SReqType = "down_prn_fmt" // with parameter csv formatted string
	SREQ_DOWN_PRN_FMT        SReqType = "down_print_format_to_scale" // with parameter csv formatted string
	SREQ_GET_AP_LIST         SReqType = "get_ap_list"
	SREQ_RESCAN_AP_LIST      SReqType = "rescan_ap_list"
	SREQ_CONNECT_AP          SReqType = "connect_ap"
	SREQ_SET_WIFI_DYNAMIC_IP SReqType = "set_wifi_dynamic_ip"
	SREQ_SET_WIFI_STATIC_IP  SReqType = "set_wifi_static_ip"
	SREQ_GET_IP_INFO         SReqType = "get_ip_info"
	SREQ_MODIFY_BT_NAME      SReqType = "modify_bt_name"
	SREQ_SEND_DATA_TO_BT     SReqType = "send_data_to_bt"
	SREQ_SEND_DATA_TO_WIFI   SReqType = "send_data_to_wifi"
	SREQ_GET_IP_MODE         SReqType = "get_ip_mode"
	SREQ_GET_WIFI_INFO       SReqType = "get_wifi_info"
	SREQ_UPDATE_FIRMWARE     SReqType = "update_firmware"
)

type ReqScaleRec struct {
	ScaleId int64
}

type ReqAddScaleRec struct {
	ScaleId int64
	Product string
	Weight  string
	Price   string
}

type ReqDelScaleRec struct {
	RecId uint
}

type ReqPrnData struct {
	ScaleModel   string   `json:"ScaleModel"`
	PrinterModel string   `json:"PrinterModel"`
	FilePaths    []string `json:"FilePaths"`
}

type ScaleRespMsg struct { // including response and unsolicited messages
	MsgType comm.RespMsgType
	MsgBody interface{} // MsgBody [T RespMsg|string]
	ScaleId int64
}

var respMsgTypeTab = []comm.RespMsgType{
	comm.WEIGHT_DATA,
	comm.ZERO_CMD_RESP,
	comm.TARE_CMD_RESP,
	comm.WEIGHT_DATA_RESP,
	comm.REG_WEIGHT_RESP,
	comm.UNREG_WEIGHT_RESP,
	comm.GET_RECS_RESP,
	comm.ADD_REC_RESP,
	comm.DEL_REC_RESP,
	comm.EN_FAC_MODE_RESP,
	comm.DIS_FAC_MODE_RESP,
	comm.EN_PASSTH_MODE_RESP,
	comm.DIS_PASSTH_MODE_RESP,
	comm.ERASE_FLASH_RESP,
	comm.WRITE_DATA_FLASH_RESP,
	comm.DOWN_PRN_FMT_RESP,
	comm.ERR_SERIAL_RESP,
	comm.GET_AP_LIST_RESP,
	comm.RESCAN_AP_LIST_RESP,
	comm.SET_WIFI_DYNAMIC_IP_RESP,
	comm.SET_WIFI_STATIC_IP_RESP,
	comm.GET_IP_INFO_RESP,
	comm.MODIFY_BT_NAME_RESP,
	comm.NO_RESP,
	comm.BT_PASSTH_DATA_RESP,
	comm.WIFI_PASSTH_DATA_RESP,
	comm.PRT_PASSTH_DATA_RESP,
	comm.UNKNOWN_DATA,
}

type RespMsg struct {
	IsAck   bool
	AckData string
}

type WeightMsg struct {
	IsStable   bool
	IsNet      bool
	WeightVal  string
	WeightUnit string
}

type RespRecs struct {
	ScaleId int64
	Recs    []ScaleRec
}
