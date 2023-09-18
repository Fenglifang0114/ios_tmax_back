package comm

type ScaleCat int

const (
	SCALE_C51 ScaleCat = iota
	SCALE_T2200
	SCALE_JWP
	SCALE_TMAX
)

type CmdType int

const (
	CMD_EN_FAC_MODE CmdType = iota
	CMD_DIS_FAC_MODE

	CMD_ZERO
	CMD_TARE
	CMD_READ_WEIGHT

	CMD_EN_CONTINUE_MODE
	CMD_DIS_CONTINUE_MODE

	CMD_EN_PASSTH
	CMD_DIS_PASSTH

	CMD_REBOOT

	CMD_WIFI_DATA_PASSTH
	CMD_WIFI_GET_AP_LIST
	CMD_WIFI_EN_DHCP
	CMD_WIFI_DIS_DHCP
	CMD_WIFI_SET_STATIC_IP
	CMD_WIFI_GET_AP_INFO
	CMD_WIFI_GET_IP_INFO
	CMD_WIFI_GET_IP_MODE
	CMD_WIFI_CONN_AP
	CMD_WIFI_DISCONN_AP

	CMD_BT_DATA_PASSTH
	CMD_MODIFY_BT_NAME

	CMD_READ_EEPROM
	CMD_WRITE_EEPROM
	CMD_ERASE_FLASH
	CMD_READ_FLASH
	CMD_WRITE_FLASH
)

type DataType int

const (
	DATA_TYPE_INT = iota
	DATA_TYPE_BYTE_ARR
	DATA_TYPE_STR
)

type CmdData struct {
	Type DataType
	Data interface{}
}

const (
	LICENSE_FILE  = "tmaxlic.txt"
	SRV_DATA_PATH = "srvdata"
)

type CmdComposer struct {
	ScaleCat   ScaleCat
	ComposeCmd func(composer *CmdComposer, cmd CmdType, cmdData CmdData) ([]byte, int, error)
}

type Packet struct {
	PayloadLen uint16
	CmdID      uint8
	CmdSubId   uint8
	SeqNum     uint8
	Payload    []byte
}

type RespMsgType string

// 处理scale回应
const (
	WEIGHT_DATA              RespMsgType = "weight_data"
	ZERO_CMD_RESP            RespMsgType = "resp_zero_cmd"
	TARE_CMD_RESP            RespMsgType = "resp_tare_cmd"
	WEIGHT_DATA_RESP         RespMsgType = "resp_weight_data"
	REG_WEIGHT_RESP          RespMsgType = "resp_reg_weight"
	UNREG_WEIGHT_RESP        RespMsgType = "resp_unreg_weight"
	GET_RECS_RESP            RespMsgType = "resp_get_recs"
	ADD_REC_RESP             RespMsgType = "resp_add_rec"
	DEL_REC_RESP             RespMsgType = "resp_del_rec"
	EN_FAC_MODE_RESP         RespMsgType = "resp_en_fac_mode"
	DIS_FAC_MODE_RESP        RespMsgType = "resp_dis_fac_mode"
	EN_PASSTH_MODE_RESP      RespMsgType = "resp_en_passth_mode"
	DIS_PASSTH_MODE_RESP     RespMsgType = "resp_dis_passth_mode"
	ERASE_FLASH_RESP         RespMsgType = "resp_erase_flash"
	WRITE_DATA_FLASH_RESP    RespMsgType = "resp_write_data_flash"
	DOWN_PRN_FMT_RESP        RespMsgType = "resp_down_prn_fmt"
	ERR_SERIAL_RESP          RespMsgType = "resp_err_serial"
	GET_AP_LIST_RESP         RespMsgType = "resp_get_ap_list"
	RESCAN_AP_LIST_RESP      RespMsgType = "resp_rescan_ap_list"
	CONNECT_AP_RESP          RespMsgType = "resp_connect_ap"
	SET_WIFI_DYNAMIC_IP_RESP RespMsgType = "resp_set_wifi_dynamic_ip"
	SET_WIFI_STATIC_IP_RESP  RespMsgType = "resp_set_wifi_static_ip"
	GET_WIFI_AP_INFO_RESP    RespMsgType = "resp_get_wifi_ap_info"
	GET_IP_INFO_RESP         RespMsgType = "resp_get_ip_info"
	GET_IP_MODE_RESP         RespMsgType = "resp_get_ip_mode"
	MODIFY_BT_NAME_RESP      RespMsgType = "resp_modify_bt_name"
	NO_RESP                  RespMsgType = "resp_no_response"
	BT_PASSTH_DATA_RESP      RespMsgType = "resp_bt_passth_data"
	WIFI_PASSTH_DATA_RESP    RespMsgType = "resp_wifi_passth_data"
	PRT_PASSTH_DATA_RESP     RespMsgType = "resp_prt_passth_data"
	SEND_DATA_TO_BT_RESP     RespMsgType = "resp_bt_passth_data"
	SEND_DATA_TO_WIFI_RESP   RespMsgType = "resp_send_data_to_wifi"
	UPDATE_FIRMWARE_RESP     RespMsgType = "resp_update_firmware"
	UPDATE_FIRMWARE_PROGRESS RespMsgType = "resp_update_firmware_progress"
	CHECK_SERIAL_PORT_RESP   RespMsgType = "resp_check_serial_port"

	UNKNOWN_DATA RespMsgType = "unknown_data"
)
