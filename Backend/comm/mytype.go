package comm

type ScaleCat int

const (
	SCALE_C51 ScaleCat = iota
	SCALE_T2200
	SCALE_T2200_PASSTH
	SCALE_JWP
	SCALE_JWP_PASSTH
	SCALE_TMAX
	SCALE_TMAX_PASSTH
)

type CmdType int

const (
	CMD_CHECK_FAC_MODE CmdType = iota
	CMD_DIS_FAC_MODE

	CMD_ZERO
	CMD_TARE
	CMD_READ_WEIGHT

	CMD_EN_CONTINUE_MODE
	CMD_DIS_CONTINUE_MODE

	CMD_EN_USR_CONT_MODE

	CMD_EN_PASSTH
	CMD_DIS_PASSTH

	CMD_GET_BUILD_INFO
	CMD_GET_SCALE_INFO
	CMD_GET_FACTORY_INFO

	CMD_REBOOT

	CMD_WIFI_DATA_PASSTH
	CMD_WIFI_GET_AP_LIST
	CMD_WIFI_EN_DHCP
	CMD_WIFI_EN_DHCP_32
	CMD_WIFI_SET_STATIC_IP
	CMD_WIFI_SET_STATIC_IP_32
	CMD_WIFI_GET_AP_INFO
	CMD_WIFI_GET_AP_INFO_32
	CMD_WIFI_GET_IP_INFO
	CMD_WIFI_GET_IP_INFO_32
	CMD_WIFI_GET_IP_MODE
	CMD_WIFI_GET_IP_MODE_32
	CMD_WIFI_CONN_AP
	CMD_WIFI_CONN_AP32
	CMD_WIFI_CONN_AP_ONE_KEY
	CMD_WIFI_DISCONN_AP
	CMD_CHANGE_WIFI_MODE
	CMD_WIFI_AT_VERSION
	CMD_WIFI_AT_MODE

	CMD_BT_DATA_PASSTH
	CMD_MODIFY_BT_NAME
	CMD_READ_EEPROM_8
	CMD_READ_EEPROM_256
	CMD_READ_EEPROM_512
	CMD_WRITE_EEPROM
	CMD_ERASE_FLASH
	CMD_ERASE_FLASH_512
	CMD_READ_FLASH
	CMD_WRITE_FLASH_8
	CMD_WRITE_FLASH_256
	CMD_WRITE_FLASH_512
	CMD_DEL_PLU
	CMD_INSERT_PLU_ADDR
	CMD_GET_PLU_HEAD
	CMD_ERASE_INSERT_PLU
	CMD_GET_WEIGHT_ERR
	CMD_GET_SCALE_TIME
	CMD_SET_SCALE_TIME
	CMD_MODIFY_VAR_VALUE
	CMD_EN_FACTORY_MODE
	CMD_GET_RANDOM_DATA
	CMD_GET_BASIC_DATA
	CMD_SET_LIMIT_TO_SCALE
	CMD_OPEN_BILL_SEND
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
	PLU_BACK_PATH = "plufiles"
	AT_VERSION    = "ESP32C3"
	IS_ESP32      = true
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
	WEIGHT_DATA                   RespMsgType = "weight_data"
	ZERO_CMD_RESP                 RespMsgType = "resp_zero_cmd"
	TARE_CMD_RESP                 RespMsgType = "resp_tare_cmd"
	WEIGHT_DATA_RESP              RespMsgType = "resp_weight_data"
	REG_WEIGHT_RESP               RespMsgType = "resp_reg_weight"
	UNREG_WEIGHT_RESP             RespMsgType = "resp_unreg_weight"
	GET_RECS_RESP                 RespMsgType = "resp_get_recs"
	ADD_REC_RESP                  RespMsgType = "resp_add_rec"
	DEL_REC_RESP                  RespMsgType = "resp_del_rec"
	EN_FAC_MODE_RESP              RespMsgType = "resp_en_fac_mode"
	DIS_FAC_MODE_RESP             RespMsgType = "resp_dis_fac_mode"
	EN_PASSTH_MODE_RESP           RespMsgType = "resp_en_passth_mode"
	DIS_PASSTH_MODE_RESP          RespMsgType = "resp_dis_passth_mode"
	ERASE_FLASH_RESP              RespMsgType = "resp_erase_flash"
	WRITE_DATA_FLASH_RESP         RespMsgType = "resp_write_data_flash"
	DOWN_PRN_FMT_RESP             RespMsgType = "resp_down_prn_fmt"
	ERR_SERIAL_RESP               RespMsgType = "resp_err_serial"
	GET_AP_LIST_RESP              RespMsgType = "resp_get_ap_list"
	RESCAN_AP_LIST_RESP           RespMsgType = "resp_rescan_ap_list"
	CONNECT_AP_RESP               RespMsgType = "resp_connect_ap"
	CONNECT_AP_ONE_KEY_RESP       RespMsgType = "resp_connect_ap_one_key"
	SET_WIFI_DYNAMIC_IP_RESP      RespMsgType = "resp_set_wifi_dynamic_ip"
	SET_WIFI_STATIC_IP_RESP       RespMsgType = "resp_set_wifi_static_ip"
	GET_AT_VERSION_RESP           RespMsgType = "resp_get_at_version"
	GET_AT_MODE_RESP              RespMsgType = "resp_get_at_mode"
	GET_WIFI_AP_INFO_RESP         RespMsgType = "resp_get_wifi_ap_info"
	GET_IP_INFO_RESP              RespMsgType = "resp_get_ip_info"
	GET_IP_MODE_RESP              RespMsgType = "resp_get_ip_mode"
	MODIFY_BT_NAME_RESP           RespMsgType = "resp_modify_bt_name"
	NO_RESP                       RespMsgType = "resp_no_response"
	BT_PASSTH_DATA_RESP           RespMsgType = "resp_bt_passth_data"
	WIFI_PASSTH_DATA_RESP         RespMsgType = "resp_wifi_passth_data"
	PRT_PASSTH_DATA_RESP          RespMsgType = "resp_prt_passth_data"
	SEND_DATA_TO_BT_RESP          RespMsgType = "resp_bt_passth_data"
	SEND_DATA_TO_WIFI_RESP        RespMsgType = "resp_send_data_to_wifi"
	UPDATE_FIRMWARE_RESP          RespMsgType = "resp_update_firmware"
	UPDATE_FIRMWARE_PROGRESS      RespMsgType = "resp_update_firmware_progress"
	DOWN_FIRMWARE_WIFI_RESP       RespMsgType = "resp_update_firmware_wifi"
	CHECK_SERIAL_PORT_RESP        RespMsgType = "resp_check_serial_port"
	GET_BUILD_INFO_RESP           RespMsgType = "resp_get_build_info"
	GET_SCALE_TIME_RESP           RespMsgType = "resp_get_scale_time"      //20240112@FLF
	SET_SCALE_TIME_RESP           RespMsgType = "resp_set_scale_time"      //20240112@FLF
	GET_ONE_EEPROM_INFO_RESP      RespMsgType = "resp_get_one_eeprom_info" //20240125@FLF
	GET_ALL_EEPROM_INFO_RESP      RespMsgType = "resp_get_all_eeprom_info" //20240125@FLF
	SET_OUTPUT_FMT_RESP           RespMsgType = "resp_set_output_fmt"
	OPEN_SCALE_PASSTHROUGH_RESP   RespMsgType = "resp_open_scale_passthrough" //20231023@FLF 连续发送的透传
	CLOSE_SCALE_PASSTHROUGH_RESP  RespMsgType = "resp_close_scale_passthrough"
	SCALE_PASSTH_DATA             RespMsgType = "scale_passth_data"
	CHANGE_SCALE_PASSTH_MODE_RESP RespMsgType = "resp_change_scale_passth_mode"
	GET_SCALE_INFO_RESP           RespMsgType = "resp_get_scale_info"
	GET_FACTORY_INFO_RESP         RespMsgType = "resp_get_factory_info"
	GET_WEIGHT_ERR_RESP           RespMsgType = "resp_get_weight_err"
	DOWN_PLU_RESP                 RespMsgType = "resp_down_plu"
	DEL_PLU_RESP                  RespMsgType = "resp_del_plu"
	INSERT_PLU_RESP               RespMsgType = "resp_insert_plu"
	GET_UI_CONF_RESP              RespMsgType = "resp_get_ui_conf"
	UPDATE_UI_CONF_RESP           RespMsgType = "resp_update_ui_conf"
	CHANGE_WIFI_MODE_RESP         RespMsgType = "resp_change_wifi_mode"
	INSERT_PLU_ADDR_RESP          RespMsgType = "resp_insert_plu_addr"
	READ_FLASH_DATA_RESP          RespMsgType = "resp_read_flash_data"
	ERASE_INSERT_PLU_RESP         RespMsgType = "resp_erase_insert_plu"
	REBOOT_RESP                   RespMsgType = "resp_reboot"
	MODIFY_EEPROM_INFO_RESP       RespMsgType = "resp_modify_eeprom_info"
	DOWN_EEPROM_INFO_RESP         RespMsgType = "resp_down_eeprom_info"
	DOWN_FACTORY_INFO_FC_RESP     RespMsgType = "resp_down_factory_info"
	DOWN_FACTORY_INFO_RESP        RespMsgType = "resp_down_factory_info_tmax"
	MODIFY_VAR_RESP               RespMsgType = "resp_modify_var_value"
	SET_SERVER_IP_RESP            RespMsgType = "resp_set_server_ip"
	GET_RANDOM_DATA_RESP          RespMsgType = "resp_get_random_data"
	EN_FACTORY_MODE_RESP          RespMsgType = "resp_en_factory_mode"
	DOWN_DEFAULT_PRN_FMT_RESP     RespMsgType = "resp_down_def_prn_fmt"
	BACKUP_DEF_SETTING_RESP       RespMsgType = "resp_backup_def_setting"
	GET_EEPROM_TO_BIN_RESP        RespMsgType = "resp_get_eeprom_to_bin"
	SET_EEPROM_FROM_BIN_RESP      RespMsgType = "resp_set_eeprom_from_bin"
	GET_BASIC_DATA_RESP           RespMsgType = "resp_get_basic_data"
	SET_LIMIT_TO_SCALE_RESP       RespMsgType = "resp_set_limit_to_scale"
	SWITCH_LIMIT_RESP             RespMsgType = "resp_switch_limit_from_scale"
	REV_DETAIl_HEAD_RESP          RespMsgType = "resp_rev_detail_head"
	REV_DETAIl_MID_RESP           RespMsgType = "resp_rev_detail_mid"
	REV_DETAIl_TAIL_RESP          RespMsgType = "resp_rev_detail_tail"
	OPEN_BILL_SEND_RESP           RespMsgType = "resp_open_bill_send"

	UNKNOWN_DATA RespMsgType = "unknown_data"
)
