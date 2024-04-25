package cmd

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	m "tmaxsrv/comm"
	l "tmaxsrv/log"
	"tmaxsrv/util"
)

var (
	EN_ENG_CMD_TMAX           []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0x05, 0xf1, 0x00, 0xf9, 0x16, 0x57, 0x5e, 0xa5, 0x5a}
	GET_RANDOM_DATA_CMD_TMAX  []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0x05, 0xf7, 0x00, 0x1C, 0xC0, 0xE8, 0xF8, 0xa5, 0x5a}
	DIS_ENG_CMD_TMAX          []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0x05, 0xf2, 0x00, 0x8b, 0xfd, 0x08, 0x8d, 0xa5, 0x5a}
	ZERO_CMD_TMAX             []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0xe1, 0x03, 0x00, 0x0c, 0xD6, 0x90, 0x0C, 0xa5, 0x5a}
	TARE_CMD_TMAX             []byte = []byte{0x5A, 0xA5, 0x00, 0x0B, 0xE1, 0x05, 0x00, 0xE9, 0x00, 0x2F, 0xAA, 0xA5, 0x5A}
	REBOOT_CMD_TMAX           []byte = []byte{0x5A, 0xA5, 0x00, 0x0B, 0x05, 0x55, 0x00, 0x1E, 0x1D, 0x98, 0x49, 0xA5, 0x5A}
	READ_WEIGHT_CMD_TMAX      []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0xe1, 0x05, 0x00, 0xac, 0x24, 0x0e, 0x03, 0xa5, 0x5a}
	EN_CONT_MODE_CMD_TMAX     []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0xe1, 0x07, 0x00, 0x49, 0xf2, 0xb1, 0xa5, 0xa5, 0x5a}
	DIS_CONT_MODE_CMD_TMAX    []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0xe1, 0x08, 0x00, 0xf4, 0x75, 0x8c, 0x8d, 0xa5, 0x5a}
	EN_PASSTH_MODE_CMD_TMAX   []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0x05, 0xf3, 0x00, 0x59, 0xE4, 0xC9, 0x51, 0xa5, 0x5a}
	DIS_PASSTH_MODE_CMD_TMAX  []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0x05, 0xf4, 0x00, 0x6E, 0x2B, 0xB7, 0x2B, 0xa5, 0x5a}
	GET_BUILD_INFO_CMD_TMAX   []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0x05, 0x56, 0x00, 0x6C, 0xF6, 0xC7, 0x9A, 0xa5, 0x5a} //20230926@FLF
	GET_SCALE_INFO_CMD_TMAX   []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0x05, 0xf0, 0x00, 0x2B, 0x0F, 0x96, 0x82, 0xa5, 0x5a} //20231101@FLF
	GET_INSERT_PLU_ADDR_TMAX  []byte = []byte{0x5A, 0xA5, 0x00, 0x0B, 0xf3, 0x02, 0x00, 0xc0, 0xf4, 0xc0, 0xae, 0xA5, 0x5A}
	GET_PLU_HEAD_TMAX         []byte = []byte{0x5a, 0xa5, 0x00, 0x11, 0xf1, 0x01, 0x00, 0x20, 0x06, 0xa0, 0x00, 0x00, 0x64, 0x6c, 0xb0, 0xb0, 0x36, 0xa5, 0x5a}
	ERASE_INSERT_PLU_TMAX     []byte = []byte{0x5A, 0xA5, 0x00, 0x0B, 0xf3, 0x03, 0x00, 0x12, 0xED, 0x01, 0x72, 0xA5, 0x5A}
	GET_SCALE_TIME_CMD_TMAX   []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0xF4, 0x02, 0x00, 0xC5, 0xFF, 0x87, 0x3B, 0xa5, 0x5a}                                     //20230112@FLF
	READ_EEPROM_256_CMD_TMAX  []byte = []byte{0x5a, 0xa5, 0x00, 0x11, 0xf1, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x97, 0xab, 0x7f, 0x0d, 0xa5, 0x5a} //20240124@FLF
	READ_EEPROM_512_CMD_TMAX  []byte = []byte{0x5a, 0xa5, 0x00, 0x11, 0xf1, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x4B, 0xC6, 0xE5, 0xBA, 0xa5, 0x5a} //20240129@FLF
	GET_FACTORY_INFO_CMD_TMAX []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0x05, 0xf6, 0x00, 0xCE, 0xD9, 0x29, 0x24, 0xa5, 0x5a}
)

func NewComposerTMAX() *m.CmdComposer {

	composer := m.CmdComposer{}
	composer.ScaleCat = m.SCALE_TMAX
	composer.ComposeCmd = ComposeCmdTMAX
	return &composer
}

func ComposeCmdTMAX(composer *m.CmdComposer, cmd m.CmdType, cmdData m.CmdData) ([]byte, int, error) {
	if composer.ScaleCat != m.SCALE_TMAX {
		return nil, CMD_TIMEOUT_IMMEDIATE, fmt.Errorf("ScaleCat is not SCALE_TMAX")
	}

	switch cmd {
	case m.CMD_CHECK_FAC_MODE:
		return EN_ENG_CMD_TMAX, CMD_TIMEOUT_VERY_SHORT_200_MS, nil
	case m.CMD_DIS_FAC_MODE:
		return DIS_ENG_CMD_TMAX, CMD_TIMEOUT_VERY_SHORT_200_MS, nil
	case m.CMD_ZERO:
		return ZERO_CMD_TMAX, CMD_TIMEOUT_VERY_SHORT_200_MS, nil
	case m.CMD_TARE:
		return TARE_CMD_TMAX, CMD_TIMEOUT_VERY_SHORT_200_MS, nil
	case m.CMD_READ_WEIGHT:
		return READ_WEIGHT_CMD_TMAX, CMD_TIMEOUT_SHORT_1500_MS, nil
	case m.CMD_EN_CONTINUE_MODE:
		return EN_CONT_MODE_CMD_TMAX, CMD_TIMEOUT_VERY_SHORT_200_MS, nil
	case m.CMD_DIS_CONTINUE_MODE:
		return DIS_CONT_MODE_CMD_TMAX, CMD_TIMEOUT_VERY_SHORT_200_MS, nil
	case m.CMD_REBOOT:
		return REBOOT_CMD_TMAX, CMD_TIMEOUT_VERY_SHORT_200_MS, nil
	case m.CMD_EN_PASSTH:
		return EN_PASSTH_MODE_CMD_TMAX, CMD_TIMEOUT_VERY_SHORT_200_MS, nil
	case m.CMD_DIS_PASSTH:
		return DIS_PASSTH_MODE_CMD_TMAX, CMD_TIMEOUT_VERY_SHORT_200_MS, nil
	case m.CMD_GET_BUILD_INFO:
		return GET_BUILD_INFO_CMD_TMAX, CMD_TIMEOUT_VERY_SHORT_200_MS, nil //20230926@FLF
	case m.CMD_GET_SCALE_TIME:
		return GET_SCALE_TIME_CMD_TMAX, CMD_TIMEOUT_VERY_SHORT_200_MS, nil //20230926@FLF
	case m.CMD_READ_EEPROM_256:
		return READ_EEPROM_256_CMD_TMAX, CMD_TIMEOUT_MEDIUM_2000_MS, nil //20240125@FLF
	case m.CMD_READ_EEPROM_512:
		return READ_EEPROM_512_CMD_TMAX, CMD_TIMEOUT_MEDIUM_2000_MS, nil //20240129@FLF
	case m.CMD_GET_SCALE_INFO:
		return GET_SCALE_INFO_CMD_TMAX, CMD_TIMEOUT_SHORT_1500_MS, nil //20231101@FLF
	case m.CMD_GET_FACTORY_INFO:
		return GET_FACTORY_INFO_CMD_TMAX, CMD_TIMEOUT_SHORT_1500_MS, nil //20240417@FLF
	case m.CMD_INSERT_PLU_ADDR:
		return GET_INSERT_PLU_ADDR_TMAX, CMD_TIMEOUT_SHORT_1500_MS, nil //20231101@FLF
	case m.CMD_GET_PLU_HEAD:
		return GET_PLU_HEAD_TMAX, CMD_TIMEOUT_SHORT_1500_MS, nil //20230103@FLF
	case m.CMD_ERASE_INSERT_PLU:
		return ERASE_INSERT_PLU_TMAX, CMD_TIMEOUT_SHORT_1500_MS, nil //20230109@FLF
	case m.CMD_GET_WEIGHT_ERR:
		addr, data := parseReadFlashTMAX(cmdData.Data.(string))
		return readFlashCmdTMAX(uint32(addr), data), CMD_TIMEOUT_SHORT_1500_MS, nil
	case m.CMD_ERASE_FLASH:
		addr := cmdData.Data.(int)
		return eraseCmdTMAX(uint32(addr)), CMD_TIMEOUT_MEDIUM_2000_MS, nil
	case m.CMD_WRITE_FLASH_256:
		addr, data := parseWrDataTMAX(cmdData.Data.(string))
		return wrDataCmdTMAX(uint32(addr), data, FILE_CHUNK_SIZE_256_TMAX), CMD_TIMEOUT_MEDIUM_2000_MS, nil
	case m.CMD_WRITE_FLASH_512:
		addr, data := parseWrDataTMAX(cmdData.Data.(string))
		return wrDataCmdTMAX(uint32(addr), data, FILE_CHUNK_SIZE_512_TMAX), CMD_TIMEOUT_MEDIUM_2000_MS, nil
	case m.CMD_WIFI_DATA_PASSTH:
		return sendDataToWifiCmdTMAX(cmdData.Data.(string)), CMD_TIMEOUT_MEDIUM_2000_MS, nil
	case m.CMD_WIFI_GET_AP_LIST:
		return getApListCmdTMAX(), CMD_TIMEOUT_LONG_20000_MS, nil
	case m.CMD_WIFI_EN_DHCP:
		return EnWifiDhcpCmdTMAX(), CMD_TIMEOUT_MEDIUM_2000_MS, nil
	case m.CMD_WIFI_DIS_DHCP:
		return DisWifiDhcpCmdTMAX(), CMD_TIMEOUT_MEDIUM_2000_MS, nil
	case m.CMD_WIFI_SET_STATIC_IP:
		fields := strings.Split(cmdData.Data.(string), ",")
		return setWifiStaticIpCmdTMAX(fields[0], fields[1], fields[2]), CMD_TIMEOUT_MEDIUM_2000_MS, nil
	case m.CMD_WIFI_GET_IP_INFO:
		return getIpInfoCmdTMAX(), CMD_TIMEOUT_LONG_20000_MS, nil
	case m.CMD_CHANGE_WIFI_MODE:
		return changeWifiModeCmdTMAX(), CMD_TIMEOUT_LONG_20000_MS, nil
	case m.CMD_WIFI_GET_AP_INFO:
		return getApInfoCmdTMAX(), CMD_TIMEOUT_LONG_20000_MS, nil
	case m.CMD_WIFI_GET_IP_MODE:
		return getIpModCmdTMAX(), CMD_TIMEOUT_LONG_20000_MS, nil
	case m.CMD_WIFI_CONN_AP:
		fields := strings.Split(cmdData.Data.(string), ",")
		return connectWifiApCmdTMAX(fields[0], fields[1], fields[2]), CMD_TIMEOUT_LONG_20000_MS, nil //ESP8266 默认超时15秒
	case m.CMD_WIFI_CONN_AP_ONE_KEY:
		fields := strings.Split(cmdData.Data.(string), ",")
		return connectWifiApCmdTMAX(fields[0], fields[1], fields[2]), CMD_TIMEOUT_MEDIUM_2000_MS, nil //ESP8266
	case m.CMD_WIFI_DISCONN_AP:
		return disconnectWifiApCmdTMAX(), CMD_TIMEOUT_LONG_20000_MS, nil
	case m.CMD_BT_DATA_PASSTH:
		return sendDataToBTCmdTMAX(cmdData.Data.(string)), CMD_TIMEOUT_MEDIUM_2000_MS, nil
	case m.CMD_MODIFY_BT_NAME:
		return modifyBTNameCmdTMAX(cmdData.Data.(string)), CMD_TIMEOUT_MEDIUM_2000_MS, nil
	case m.CMD_DEL_PLU:
		return getDelPluCmdTMAX(cmdData.Data.(string)), CMD_TIMEOUT_MEDIUM_2000_MS, nil
	case m.CMD_WRITE_EEPROM:
		addr, data := parseWrDataTMAX(cmdData.Data.(string))
		return getModifyEepromCmdTMAX(uint32(addr), data), CMD_TIMEOUT_MEDIUM_2000_MS, nil
	case m.CMD_SET_SCALE_TIME:
		return getSetScaleTimeCmdTMAX(cmdData.Data.(string)), CMD_TIMEOUT_MEDIUM_2000_MS, nil
	case m.CMD_MODIFY_VAR_VALUE:
		data, _ := hex.DecodeString(cmdData.Data.(string))
		return getModifyVarValueCmdTMAX(data), CMD_TIMEOUT_MEDIUM_2000_MS, nil
	case m.CMD_EN_FACTORY_MODE:
		data, _ := hex.DecodeString(cmdData.Data.(string))
		return enFactoryModeCmdTMAX(data), CMD_TIMEOUT_VERY_SHORT_200_MS, nil
	case m.CMD_GET_RANDOM_DATA:
		return GET_RANDOM_DATA_CMD_TMAX, CMD_TIMEOUT_VERY_SHORT_200_MS, nil

	}
	return nil, CMD_TIMEOUT_IMMEDIATE, nil
}

var GET_AP_LIST_CMD []byte = []byte("AT+CWLAP\r\n")
var CONNECT_AP_CMD []byte = []byte("AT+CWJAP_DEF=\"%s\",\"%s\"\r\n") // ssid, password, bssid
var DISCONNECT_AP_CMD []byte = []byte("AT+CWQAP\r\n")                // ssid, password, bssid
var GET_AP_INFO_CMD []byte = []byte("AT+CWJAP_DEF?\r\n")             // ssid, password, bssid
var GET_IP_INFO_CMD []byte = []byte("AT+CIPSTA_CUR?\r\n")            // ssid, password, bssid
var GET_IP_MODE_CMD []byte = []byte("AT+CWDHCP_CUR?\r\n")
var EN_DHCP_DEF_CMD []byte = []byte("AT+CWDHCP_DEF=1,1\r\n")
var DIS_DHCP_DEF_CMD []byte = []byte("AT+CWDHCP_DEF=1,0")
var SET_WIFI_STATIC_IP_DEF_CMD []byte = []byte("AT+CIPSTA_DEF=\"%s\",\"%s\",\"%s\"\r\n") // ip, gateway, netmask
var CHANGE_WIFI_MODE_CMD []byte = []byte("AT+CWMODE=1\r\n")

const (
	PACKET_HEAD_TMAX         = 0x5AA5
	PACKET_TAIL_TMAX         = 0xA55A
	FILE_CHUNK_SIZE_256_TMAX = 275 //FLF
	FILE_CHUNK_SIZE_512_TMAX = 531 //@FLF20240110
	EARSE_CHUNK_SIZE_TMAX    = 0x13
	ERASE_SIZE_TMAX          = 0x800
)

const (
	CMDID_READ_WEIGHT_TMAX         = 0xE101
	CMDID_READ_STABLE_WEIGHT_TMAX  = 0xE102
	CMDID_PREF_ZERO_TMAX           = 0xE103
	CMDID_PREF_ZERO_ON_STABLE_TMAX = 0xE104
	CMDID_PREF_TARE_TMAX           = 0xE105
	CMDID_PREF_TARE_ON_STABLE_TMAX = 0xE106
	CMDID_EN_CONT_WEIGHT_TMAX      = 0xE107
	CMDID_DIS_CONT_WEIGHT_TMAX     = 0xE108
	CMDID_SET_1ST_CAP_TMAX         = 0xE109
	CMDID_SET_2ND_CAP_TMAX         = 0xE10A
)

const (
	CMDID_SEND_DATA_TO_BT_TMAX   = 0xF201
	CMDID_SEND_DATA_TO_WIFI_TMAX = 0xF202
	CMDID_SEND_DATA_TO_PRN_TMAX  = 0xF203
)

const (
	CMDID_REBOOT_TMAX            = 0x0555
	CMDID_READ_SCALE_INFO_TMAX   = 0x05F0
	CMDID_EN_FAC_TMAX            = 0x05F1
	CMDID_DIS_FAC_TMAX           = 0x05F2
	CMDID_EN_PASSTH_TMAX         = 0x05F3
	CMDID_DIS_PASSTH_TMAX        = 0x05F4
	CMDID_GET_MAX_PACK_SIZE_TMAX = 0x05F5
	CMDID_GET_FACTORY_INFO_TMAX  = 0x05F6
	CMDID_GET_RANDOM_DATA        = 0x05F7
)
const (
	CMDID_READ_FLASH_TMAX        = 0xF101
	CMDID_WRITE_FLASH_TMAX       = 0xF102
	CMDID_ERASE_FLASH_TMAX       = 0xF103
	CMDID_READ_EEPROM_TMAX       = 0xF104
	CMDID_WRITE_EEPROM_TMAX      = 0xF105
	CMDID_ERASE_EEPROM_TMAX      = 0xF106
	CMDID_READ_ROM_TMAX          = 0xF107
	CMDID_WRITE_ROM_TMAX         = 0xF108
	CMDID_ERASE_ROM_TMAX         = 0xF109
	CMDID_DEL_PLU_TMAX           = 0xF301
	CMDID_INSERT_PLU_TMAX        = 0xF302
	CMDID_ERASE_INSERT_PLU_TMAX  = 0xF303
	CMDID_SET_SCALE_TIME_TMAX    = 0xF401
	CMDID_GET_SCALE_TIME_TMAX    = 0xF402
	CMDID_SCALE_PASSTH_DATA_TMAX = 0xFF23 // virtual command ID
	CMDID_DOWN_PLU_TMAX          = 0xFF24
	CMDID_MODIFY_VAR_TMAX        = 0xF501
	CMDID_EN_FACTORY_MODE        = 0xF601
)

const (
	RESP_RESULT_OK   = 0x06
	RESP_RESULT_FAIL = 0x15
)

const (
	CMD_IDENTIFY_IDX = 2
	CMD_TYPE_IDX     = 3
)

const PACK_LEN_WITHOUT_DATA_TMAX = 13 // for TMAX protocol version, 13: Head: 2 + len:2 + CmdId: 1 + SubCmdId: 1 + SeqNo:1 + CRC:4 + Tail:2
func composeCmd(cmdID uint16, seqNo byte, data []byte) []byte {
	// Calculate packet length
	var packLen uint16 = uint16(PACK_LEN_WITHOUT_DATA_TMAX + len(data))
	cmd := make([]byte, packLen)
	binary.BigEndian.PutUint16(cmd[0:2], PACKET_HEAD_TMAX)
	// 构建命令ID与命令类型
	binary.BigEndian.PutUint16(cmd[2:4], packLen-2) // packet length without header
	binary.BigEndian.PutUint16(cmd[4:6], cmdID)
	cmd[6] = seqNo
	if len(data) > 0 {
		copy(cmd[7:], data)
	}
	// 计算与添加校验码
	checksum := util.Crc32MPEG2(cmd[2 : packLen-6])
	binary.BigEndian.PutUint32(cmd[packLen-6:], checksum)
	// 添加包尾
	binary.BigEndian.PutUint16(cmd[packLen-2:], PACKET_TAIL_TMAX)
	return cmd
}

const PACK_LEN_HEADER_TMAX = 13 //9 = 2个头+2个f501+1个00+2个长度+4个校验位+2个尾巴
// 构建命令   //FLF
func composeModifyVarCmd(cmdID uint16, data []byte) []byte {
	// Calculate packet length
	var packLen uint16 = uint16(PACK_LEN_HEADER_TMAX + len(data))
	cmd := make([]byte, packLen)
	binary.BigEndian.PutUint16(cmd[0:2], PACKET_HEAD_TMAX)
	// 构建命令ID与命令类型
	binary.BigEndian.PutUint16(cmd[2:4], packLen-2) // packet length without header
	binary.BigEndian.PutUint16(cmd[4:6], cmdID)
	cmd[6] = 0x00
	if len(data) > 0 {
		copy(cmd[7:], data)
	}
	// 计算与添加校验码
	checksum := util.Crc32MPEG2(cmd[2 : packLen-6])
	binary.BigEndian.PutUint32(cmd[packLen-6:], checksum)
	// 添加包尾
	binary.BigEndian.PutUint16(cmd[packLen-2:], PACKET_TAIL_TMAX)
	fmt.Printf("%x", cmd)

	return cmd
}

const PACK_LEN_FACTORY_TMAX = 17 // = 2个头+2个长度+2个f5+1个00+4个数据+4个校验位+2个尾巴
// 构建命令   //FLF
func composeFactoryCmd(cmdID uint16, data []byte) []byte {
	// Calculate packet length
	var packLen uint16 = uint16(PACK_LEN_FACTORY_TMAX)
	cmd := make([]byte, packLen)
	binary.BigEndian.PutUint16(cmd[0:2], PACKET_HEAD_TMAX)
	// 构建命令ID与命令类型
	binary.BigEndian.PutUint16(cmd[2:4], packLen-2) // packet length without header
	binary.BigEndian.PutUint16(cmd[4:6], cmdID)
	cmd[6] = 0x00
	//data 有6个字节，data[0]data[1] 随机数 data[2]data[3]data[4]data[5] md5
	//data[0]和data[2]做XOR  得到的数据x1
	//data[1]和data[5]做XOR  得到的数据x2
	//x1 data[3] data[4] x2  4个字节做CRC得到此命令的数据部分
	data[2] = data[0] ^ data[2]
	data[5] = data[1] ^ data[5]
	dataCrc := util.Crc32MPEG2(data[2:]) //此处是数据，只不过数据值是CRC的校验值，后面还是有校验的
	binary.BigEndian.PutUint32(cmd[7:11], dataCrc)
	// 计算与添加校验码
	checksum := util.Crc32MPEG2(cmd[2 : packLen-6])
	binary.BigEndian.PutUint32(cmd[packLen-6:], checksum)
	// 添加包尾
	binary.BigEndian.PutUint16(cmd[packLen-2:], PACKET_TAIL_TMAX)
	fmt.Printf("%x", cmd)

	return cmd
}

// 构建一个数据包   //FLF
func wrDataCmdTMAX(addr uint32, data []byte, packetLen uint16) []byte {
	// 构建包头
	packet := make([]byte, packetLen)
	binary.BigEndian.PutUint16(packet[0:2], PACKET_HEAD_TMAX)
	//构建数据长度  总长度-包头
	binary.BigEndian.PutUint16(packet[2:4], packetLen-2)

	// 构建命令ID与命令类型
	binary.BigEndian.PutUint16(packet[4:6], uint16(CMDID_WRITE_FLASH_TMAX))
	//构建保留数据 00
	binary.BigEndian.PutUint16(packet[6:8], 0x00)
	// 构建地址
	binary.BigEndian.PutUint32(packet[7:11], addr)

	// 构建数据长度
	binary.BigEndian.PutUint16(packet[11:13], uint16(len(data)))

	// 复制数据
	copy(packet[13:], data)

	// 计算与添加校验码
	checksum := util.Crc32MPEG2(packet[2 : packetLen-6])
	binary.BigEndian.PutUint32(packet[packetLen-6:], checksum)

	// 添加包尾
	binary.BigEndian.PutUint16(packet[packetLen-2:], PACKET_TAIL_TMAX)

	return packet
}

// CMD:  5a a5 00 11 f1 01 00 20 04 C0 00 00 08 5C F9 F4 7B a5 5a   //FLF
// 读取flash组命令
func readFlashCmdTMAX(addr uint32, data []byte) []byte { // erase size will 2K
	// 构建包头
	packet := make([]byte, EARSE_CHUNK_SIZE_TMAX)
	binary.BigEndian.PutUint16(packet[0:2], PACKET_HEAD_TMAX)
	//构建数据长度  总长度-包头
	binary.BigEndian.PutUint16(packet[2:4], EARSE_CHUNK_SIZE_TMAX-2)
	// 构建命令ID与命令类型
	binary.BigEndian.PutUint16(packet[4:6], uint16(CMDID_READ_FLASH_TMAX))
	//构建保留数据 00
	binary.BigEndian.PutUint16(packet[6:8], 0x00)
	// 构建地址
	binary.BigEndian.PutUint32(packet[7:11], addr)
	// 读取长度
	copy(packet[11:13], data)
	// 计算与添加校验码
	checksum := util.Crc32MPEG2(packet[2 : EARSE_CHUNK_SIZE_TMAX-6])
	binary.BigEndian.PutUint32(packet[EARSE_CHUNK_SIZE_TMAX-6:], checksum)

	// 添加包尾
	binary.BigEndian.PutUint16(packet[EARSE_CHUNK_SIZE_TMAX-2:], PACKET_TAIL_TMAX)

	return packet
}

// CMD:  5a a5 00 11 f1 03 00 08 01 e0 00 08 00 75 E8 A3 E3 a5 5a   //FLF
// 擦除原本秤上的打印格式
func eraseCmdTMAX(addr uint32) []byte { // erase size will 2K
	// 构建包头
	packet := make([]byte, EARSE_CHUNK_SIZE_TMAX)
	binary.BigEndian.PutUint16(packet[0:2], PACKET_HEAD_TMAX)
	//构建数据长度  总长度-包头
	binary.BigEndian.PutUint16(packet[2:4], EARSE_CHUNK_SIZE_TMAX-2)
	// 构建命令ID与命令类型
	binary.BigEndian.PutUint16(packet[4:6], uint16(CMDID_ERASE_FLASH_TMAX))
	//构建保留数据 00
	binary.BigEndian.PutUint16(packet[6:8], 0x00)
	// 构建地址
	binary.BigEndian.PutUint32(packet[7:11], addr)
	// 擦除长度
	binary.BigEndian.PutUint16(packet[11:13], ERASE_SIZE_TMAX)

	// 计算与添加校验码
	checksum := util.Crc32MPEG2(packet[2 : EARSE_CHUNK_SIZE_TMAX-6])
	binary.BigEndian.PutUint32(packet[EARSE_CHUNK_SIZE_TMAX-6:], checksum)

	// 添加包尾
	binary.BigEndian.PutUint16(packet[EARSE_CHUNK_SIZE_TMAX-2:], PACKET_TAIL_TMAX)

	return packet
}

func MyStringToBytes(str string) []byte {
	byteSlice := make([]byte, len(str))
	copy(byteSlice, str)
	return byteSlice
}

// 修改蓝牙名称
func modifyBTNameCmdTMAX(name string) []byte {
	l.Log.Debug("compose modify BT name cmd")
	// MODIFY_BT_NAME_CHUNK_SIZE should include all data except BT name
	data := "TTM:REN-" + name + "\r\n\x00"
	bytes := MyStringToBytes(data)
	return composeCmd(0xf201, 0, bytes)
}

// Get AP list
func getApListCmdTMAX() []byte {
	l.Log.Debug("compose Get AP list cmd")
	return composeCmd(0xf202, 0, GET_AP_LIST_CMD)
}

// Change Wifi Mode
func changeWifiModeCmdTMAX() []byte {
	l.Log.Debug("compose change wifi mode cmd")
	return composeCmd(0xf202, 0, CHANGE_WIFI_MODE_CMD)
}

// Send data to BT
func sendDataToBTCmdTMAX(data string) []byte {
	l.Log.Debug("compose send data to BT cmd")
	return composeCmd(0xf201, 0, []byte(data))
}

// Send data to Wifi
func sendDataToWifiCmdTMAX(data string) []byte {
	l.Log.Debug("compose send data to Wifi cmd")
	return composeCmd(0xf202, 0, []byte(data))
}

func EnWifiDhcpCmdTMAX() []byte {
	l.Log.Debug("compose set wifi dynamic IP cmd")
	return composeCmd(0xf202, 0, EN_DHCP_DEF_CMD)
}

func DisWifiDhcpCmdTMAX() []byte {
	l.Log.Debug("compose disable wifi dhcp cmd")
	return composeCmd(0xf202, 0, DIS_DHCP_DEF_CMD)
}

func setWifiStaticIpCmdTMAX(ip string, gateway string, netmask string) []byte {
	l.Log.Debug("compose set wifi to static IP cmd")
	return composeCmd(0xf202, 0, []byte(fmt.Sprintf(string(SET_WIFI_STATIC_IP_DEF_CMD), ip, gateway, netmask)))
}

// Connect to specifi AP
func connectWifiApCmdTMAX(ssid string, bssid string, passwd string) []byte {
	l.Log.Debug("compose connect to Wifi AP cmd")
	return composeCmd(0xf202, 0, []byte(fmt.Sprintf(string(CONNECT_AP_CMD), ssid, passwd)))
}

// Connect to specifi AP
func disconnectWifiApCmdTMAX() []byte {
	l.Log.Debug("compose disconnect to Wifi AP cmd")
	return composeCmd(0xf202, 0, []byte{})
}

// Get wifi AP info from scale
func getApInfoCmdTMAX() []byte {
	l.Log.Debug("compose get IP info cmd")
	return composeCmd(0xf202, 0, GET_AP_INFO_CMD)
}

// Get IP info from scale
func getIpInfoCmdTMAX() []byte {
	l.Log.Debug("compose get IP info cmd")
	return composeCmd(0xf202, 0, GET_IP_INFO_CMD)
}

func getIpModCmdTMAX() []byte {
	l.Log.Debug("compose get IP mode cmd")
	return composeCmd(0xf202, 0, GET_IP_MODE_CMD)
}

func getDelPluCmdTMAX(data string) []byte {
	l.Log.Debug("compose get IP mode cmd")
	return composeCmd(0xf301, 0, []byte(data))
}

func getModifyEepromCmdTMAX(addr uint32, data []byte) []byte {
	l.Log.Debug("compose modify eeprom info cmd")
	var dataLen = 19 + len(data)
	return wrDataCmdTMAX(addr, data, uint16(dataLen))
}

func getModifyVarValueCmdTMAX(data []byte) []byte {
	l.Log.Debug("compose modify var value cmd")
	return composeModifyVarCmd(0xf501, data)
}

func enFactoryModeCmdTMAX(data []byte) []byte {
	l.Log.Debug("compose en factory mode cmd")
	return composeFactoryCmd(0xf601, data)
}

func getSetScaleTimeCmdTMAX(data string) []byte {
	l.Log.Debug("compose get IP mode cmd")
	return composeCmd(0xf401, 0, []byte(data))
}

func parseWrDataTMAX(inData string) (addr int64, data []byte) { // inData is hex ascii
	addr, err := strconv.ParseInt(inData[0:8], 16, 64)
	if err != nil {
		fmt.Println("Invalid offset number")
		return -1, nil
	}

	// Extract byte array
	data, err = hex.DecodeString(inData[9:])
	if err != nil {
		fmt.Println("Invalid hex string")
		return -1, nil
	}

	return
}

func parseReadFlashTMAX(inData string) (addr int64, data []byte) { // inData is hex ascii
	addr, err := strconv.ParseInt(inData[0:8], 16, 64)
	if err != nil {
		fmt.Println("Invalid offset number")
		return -1, nil
	}

	// Extract byte array
	data, err = hex.DecodeString(inData[9:])
	if err != nil {
		fmt.Println("Invalid hex string")
		return -1, nil
	}

	return
}
