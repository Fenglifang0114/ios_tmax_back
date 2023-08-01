package svc

import (
	"net"

	"go.bug.st/serial"
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
	MediaConf MediaConf
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
	SREQ_DOWN_PRN_FMT          SReqType = "down_print_format_to_scale" // with parameter csv formatted string
	SREQ_GET_AP_LIST           SReqType = "get_ap_list"
	SREQ_RESCAN_AP_LIST        SReqType = "rescan_ap_list"
	SREQ_CONNECT_AP_DYNAMIC_IP SReqType = "connect_ap_dynamic_ip"
	SREQ_CONNECT_AP_STATIC_IP  SReqType = "connect_ap_static_ip"
	SREQ_GET_IP_INFO           SReqType = "get_ip_info"
	SREQ_MODIFY_BT_NAME        SReqType = "modify_bt_name"
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

type ScaleRespMsg struct { // including response and unsolicited messages
	MsgType RespMsgType
	MsgBody interface{} // MsgBody [T RespMsg|string]
	ScaleId int64
}

type RespMsgType string

const (
	WEIGHT_DATA                RespMsgType = "weight_data"
	ZERO_CMD_RESP              RespMsgType = "resp_zero_cmd"
	TARE_CMD_RESP              RespMsgType = "resp_tare_cmd"
	WEIGHT_DATA_RESP           RespMsgType = "resp_weight_data"
	REG_WEIGHT_RESP            RespMsgType = "resp_reg_weight"
	UNREG_WEIGHT_RESP          RespMsgType = "resp_unreg_weight"
	GET_RECS_RESP              RespMsgType = "resp_get_recs"
	ADD_REC_RESP               RespMsgType = "resp_add_rec"
	DEL_REC_RESP               RespMsgType = "resp_del_rec"
	EN_FAC_MODE_RESP           RespMsgType = "resp_en_fac_mode"
	DIS_FAC_MODE_RESP          RespMsgType = "resp_dis_fac_mode"
	EN_PASSTH_MODE_RESP        RespMsgType = "resp_en_passth_mode"
	DIS_PASSTH_MODE_RESP       RespMsgType = "resp_dis_passth_mode"
	ERASE_FLASH_RESP           RespMsgType = "resp_erase_flash"
	WRITE_DATA_FLASH_RESP      RespMsgType = "resp_write_data_flash"
	DOWN_PRN_FMT_RESP          RespMsgType = "resp_down_prn_fmt"
	ERR_SERIAL_RESP            RespMsgType = "resp_err_serial"
	GET_AP_LIST_RESP           RespMsgType = "resp_get_ap_list"
	RESCAN_AP_LIST_RESP        RespMsgType = "resp_rescan_ap_list"
	CONNECT_AP_DYNAMIC_IP_RESP RespMsgType = "resp_connect_ap_dynamic_ip"
	CONNECT_AP_STATIC_IP_RESP  RespMsgType = "resp_connect_ap_static_ip"
	GET_IP_INFO_RESP           RespMsgType = "resp_get_ip_info"
	MODIFY_BT_NAME_RESP        RespMsgType = "resp_modify_bt_name"
	NO_RESP                    RespMsgType = "resp_no_response"
	BT_PASSTH_DATA             RespMsgType = "bt_passth_data"
	WIFI_PASSTH_DATA           RespMsgType = "wifi_passth_data"
	PRT_PASSTH_DATA            RespMsgType = "prt_passth_data"
	UNKNOWN_DATA               RespMsgType = "unknown_data"
)

var respMsgTypeTab = []RespMsgType{
	WEIGHT_DATA,
	ZERO_CMD_RESP,
	TARE_CMD_RESP,
	WEIGHT_DATA_RESP,
	REG_WEIGHT_RESP,
	UNREG_WEIGHT_RESP,
	GET_RECS_RESP,
	ADD_REC_RESP,
	DEL_REC_RESP,
	EN_FAC_MODE_RESP,
	DIS_FAC_MODE_RESP,
	EN_PASSTH_MODE_RESP,
	DIS_PASSTH_MODE_RESP,
	ERASE_FLASH_RESP,
	WRITE_DATA_FLASH_RESP,
	DOWN_PRN_FMT_RESP,
	ERR_SERIAL_RESP,
	GET_AP_LIST_RESP,
	RESCAN_AP_LIST_RESP,
	CONNECT_AP_DYNAMIC_IP_RESP,
	CONNECT_AP_STATIC_IP_RESP,
	GET_IP_INFO_RESP,
	MODIFY_BT_NAME_RESP,
	NO_RESP,
	BT_PASSTH_DATA,
	WIFI_PASSTH_DATA,
	PRT_PASSTH_DATA,
	UNKNOWN_DATA,
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
