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
	EN_ENG_CMD_TMAX          []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0x05, 0xf1, 0x00, 0xf9, 0x16, 0x57, 0x5e, 0xa5, 0x5a}
	DIS_ENG_CMD_TMAX         []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0x05, 0xf2, 0x00, 0x8b, 0xfd, 0x08, 0x8d, 0xa5, 0x5a}
	ZERO_CMD_TMAX            []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0xe1, 0x03, 0x00, 0x0c, 0xD6, 0x90, 0x0C, 0xa5, 0x5a}
	TARE_CMD_TMAX            []byte = []byte{0x5A, 0xA5, 0x00, 0x0B, 0xE1, 0x05, 0x00, 0xE9, 0x00, 0x2F, 0xAA, 0xA5, 0x5A}
	REBOOT_CMD_TMAX          []byte = []byte{} // contact MCU scale engineer
	READ_WEIGHT_CMD_TMAX     []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0xe1, 0x05, 0x00, 0xac, 0x24, 0x0e, 0x03, 0xa5, 0x5a}
	EN_CONT_MODE_CMD_TMAX    []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0xe1, 0x07, 0x00, 0x49, 0xf2, 0xb1, 0xa5, 0xa5, 0x5a}
	DIS_CONT_MODE_CMD_TMAX   []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0xe1, 0x08, 0x00, 0xf4, 0x75, 0x8c, 0x8d, 0xa5, 0x5a}
	EN_PASSTH_MODE_CMD_TMAX  []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0x05, 0xf3, 0x00, 0x59, 0xE4, 0xC9, 0x51, 0xa5, 0x5a}
	DIS_PASSTH_MODE_CMD_TMAX []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0x05, 0xf4, 0x00, 0x6E, 0x2B, 0xB7, 0x2B, 0xa5, 0x5a}
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
	case m.CMD_EN_FAC_MODE:
		return EN_ENG_CMD_TMAX, CMD_TIMEOUT_SHORT_100_MS, nil
	case m.CMD_DIS_FAC_MODE:
		return DIS_ENG_CMD_TMAX, CMD_TIMEOUT_SHORT_100_MS, nil
	case m.CMD_ZERO:
		return ZERO_CMD_TMAX, CMD_TIMEOUT_SHORT_100_MS, nil
	case m.CMD_TARE:
		return TARE_CMD_TMAX, CMD_TIMEOUT_SHORT_100_MS, nil
	case m.CMD_READ_WEIGHT:
		return READ_WEIGHT_CMD_TMAX, CMD_TIMEOUT_SHORT_100_MS, nil
	case m.CMD_EN_CONTINUE_MODE:
		return EN_CONT_MODE_CMD_TMAX, CMD_TIMEOUT_SHORT_100_MS, nil
	case m.CMD_DIS_CONTINUE_MODE:
		return DIS_CONT_MODE_CMD_TMAX, CMD_TIMEOUT_SHORT_100_MS, nil
	case m.CMD_REBOOT:
		return REBOOT_CMD_TMAX, CMD_TIMEOUT_SHORT_100_MS, nil
	case m.CMD_EN_PASSTH:
		return EN_PASSTH_MODE_CMD_TMAX, CMD_TIMEOUT_SHORT_100_MS, nil
	case m.CMD_DIS_PASSTH:
		return DIS_PASSTH_MODE_CMD_TMAX, CMD_TIMEOUT_SHORT_100_MS, nil
	case m.CMD_ERASE_FLASH:
		addr := cmdData.Data.(int)
		return eraseCmdTMAX(uint32(addr)), CMD_TIMEOUT_MEDIUM_2000_MS, nil
	case m.CMD_WRITE_FLASH:
		addr, data := parseWrDataTMAX(cmdData.Data.(string))
		return wrDataCmdTMAX(uint32(addr), data), CMD_TIMEOUT_MEDIUM_2000_MS, nil
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
	case m.CMD_WIFI_GET_AP_INFO:
		return getApInfoCmdTMAX(), CMD_TIMEOUT_LONG_20000_MS, nil
	case m.CMD_WIFI_GET_IP_MODE:
		return getIpModCmdTMAX(), CMD_TIMEOUT_LONG_20000_MS, nil
	case m.CMD_WIFI_CONN_AP:
		fields := strings.Split(cmdData.Data.(string), ",")
		return connectWifiApCmdTMAX(fields[0], fields[1], fields[2]), CMD_TIMEOUT_LONG_20000_MS, nil
	case m.CMD_WIFI_DISCONN_AP:
		return disconnectWifiApCmdTMAX(), CMD_TIMEOUT_LONG_20000_MS, nil
	case m.CMD_BT_DATA_PASSTH:
		return sendDataToBTCmdTMAX(cmdData.Data.(string)), CMD_TIMEOUT_MEDIUM_2000_MS, nil
	case m.CMD_MODIFY_BT_NAME:
		return modifyBTNameCmdTMAX(cmdData.Data.(string)), CMD_TIMEOUT_MEDIUM_2000_MS, nil
	}
	return nil, CMD_TIMEOUT_IMMEDIATE, nil
}

var GET_AP_LIST_CMD []byte = []byte("AT+CWLAP\r\n")
var CONNECT_AP_CMD []byte = []byte("AT+CWJAP_DEF=\"%s\",\"%s\"\r\n") // ssid, password, bssid
var DISCONNECT_AP_CMD []byte = []byte("AT+CWQAP\r\n")                // ssid, password, bssid
var GET_AP_INFO_CMD []byte = []byte("AT+CWJAP_DEF?\r\n")             // ssid, password, bssid
var GET_IP_INFO_CMD []byte = []byte("AT+CIPSTA_CUR?\r\n")            // ssid, password, bssid
var GET_IP_MODE_CMD []byte = []byte("AT+CWDHCP_CUR?\r\n")            // FIXME:
var EN_DHCP_DEF_CMD []byte = []byte("AT+CWDHCP_DEF=1,1\r\n")
var DIS_DHCP_DEF_CMD []byte = []byte("AT+CWDHCP_DEF=1,0")
var SET_WIFI_STATIC_IP_DEF_CMD []byte = []byte("AT+CIPSTA_DEF=\"%s\",\"%s\",\"%s\"\r\n") // ip, gateway, netmask

const (
	FLASH_ADDR                     = 0x08003000
	PACKET_HEAD_TMAX               = 0x5AA5
	CMD_IDENTIFY_TMAX              = 0xA0
	CMD_TYPE                       = 0xB0
	PACKET_TAIL_TMAX               = 0xA55A
	DATA_LENGTH                    = 256
	FILE_CHUNK_SIZE_TMAX           = 272
	OPEN_FAC_CHUNK_SIZE_TMAX       = 6
	CLOSE_FAC_CHUNK_SIZE_TMAX      = 6
	EN_PASSTH_CHUNK_SIZE_TMAX      = 6
	DIS_PASSTH_CHUNK_SIZE_TMAX     = 6
	MODIFY_BT_NAME_CHUNK_SIZE_TMAX = 24
	REC_CHUNK_SIZE_TMAX            = 11
	EARSE_CHUNK_SIZE_TMAX          = 0x10
	CMD_ERASE_TMAX                 = 0xA2
	CMD_ERASE_SIZE_TMAx            = 0x800
	CMD_FLASH_TMAX                 = 0xB0
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
	binary.BigEndian.PutUint16(cmd[0:2], PACKET_HEAD)
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
	binary.BigEndian.PutUint16(cmd[packLen-2:], PACKET_TAIL)

	return cmd
}

// 构建一个数据包
func wrDataCmdTMAX(addr uint32, data []byte) []byte {
	// 构建包头
	packet := make([]byte, FILE_CHUNK_SIZE)
	binary.BigEndian.PutUint16(packet[0:2], PACKET_HEAD)

	// 构建命令ID与命令类型
	packet[2] = CMD_IDENTIFY
	packet[3] = CMD_TYPE

	// 构建地址
	binary.BigEndian.PutUint32(packet[4:8], addr)

	// 构建数据长度
	binary.BigEndian.PutUint16(packet[8:10], uint16(len(data)))

	// 复制数据
	copy(packet[10:], data)

	// 计算与添加校验码
	checksum := util.Crc32MPEG2(packet[2 : FILE_CHUNK_SIZE-6])
	binary.BigEndian.PutUint32(packet[FILE_CHUNK_SIZE-6:], checksum)

	// 添加包尾
	binary.BigEndian.PutUint16(packet[FILE_CHUNK_SIZE-2:], PACKET_TAIL)

	return packet
}

// 擦除原本秤上的打印格式
func eraseCmdTMAX(addr uint32) []byte { // erase size will 2K
	// 构建包头
	packet := make([]byte, EARSE_CHUNK_SIZE)
	binary.BigEndian.PutUint16(packet[0:2], PACKET_HEAD)
	// 构建命令ID与命令类型
	packet[2] = CMD_ERASE
	packet[3] = CMD_FLASH

	// 构建地址
	binary.BigEndian.PutUint32(packet[4:8], addr)

	// 擦除长度
	binary.BigEndian.PutUint16(packet[8:10], CMD_ERASE_SIZE)

	// 计算与添加校验码
	checksum := util.Crc32MPEG2(packet[2 : EARSE_CHUNK_SIZE-6])
	binary.BigEndian.PutUint32(packet[EARSE_CHUNK_SIZE-6:], checksum)

	// 添加包尾
	binary.BigEndian.PutUint16(packet[EARSE_CHUNK_SIZE-2:], PACKET_TAIL)

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
