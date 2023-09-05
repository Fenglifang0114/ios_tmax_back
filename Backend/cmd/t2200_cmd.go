package cmd

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"

	m "tmaxsrv/comm"
	"tmaxsrv/util"
)

const (
	//FLASH_ADDR          = 0x08003000
	PACKET_HEAD  = 0x5AA5
	CMD_IDENTIFY = 0xA0
	// CMD_TYPE            = 0xB0
	PACKET_TAIL = 0xA55A
	// DATA_LENGTH         = 256
	FILE_CHUNK_SIZE     = 272
	OPEN_FAC_CHUNK_SIZE = 6
	REC_CHUNK_SIZE      = 11
	EARSE_CHUNK_SIZE    = 0x10
	CMD_ERASE           = 0xA2
	CMD_ERASE_SIZE      = 0x800
	CMD_FLASH           = 0xB0
)

const (
	READ_WEIGHT_CMD_T2200 string = "W\r\n"
	TARE_CMD_T2200        string = "T\r\n"
	ZERO_CMD_T2200        string = "Z\r\n"
)

var (
	REBOOT_CMD_T2200 []byte = []byte{0x5A, 0xA5, 0x05, 0x55, 0xA5, 0x5A}
)

func NewComposerT2200() *m.CmdComposer {

	composer := m.CmdComposer{}
	composer.ScaleCat = m.SCALE_T2200
	composer.ComposeCmd = ComposeCmdT2200

	return &composer
}

func ComposeCmdT2200(composer *m.CmdComposer, cmd m.CmdType, cmdData m.CmdData) ([]byte, int, error) {
	if composer.ScaleCat != m.SCALE_T2200 {
		return nil, -1, fmt.Errorf("ScaleCat is not SCALE_T2200")
	}

	switch cmd {
	case m.CMD_EN_FAC_MODE:
		return enFacModeCmdT2200()
	case m.CMD_DIS_FAC_MODE:
		return disFacModeCmdT2200()
	case m.CMD_ZERO:
		return zeroCmdT2200()
	case m.CMD_TARE:
		return tareCmdT2200()
	case m.CMD_READ_WEIGHT:
		return readWeightCmdT2200()
	case m.CMD_REBOOT:
		return rebootCmdT2200()
	case m.CMD_ERASE_FLASH:
		addr := cmdData.Data.(int)
		return eraseCmdT2200(uint32(addr))
	case m.CMD_WRITE_FLASH:
		addr, data := parseWrDataT2200(cmdData.Data.(string))
		return wrDataCmdT2200(uint32(addr), data)
	}
	return nil, -1, nil
}

func zeroCmdT2200() ([]byte, int, error) {
	return []byte(ZERO_CMD_T2200), CMD_TIMEOUT_IMMEDIATE, nil
}

func tareCmdT2200() ([]byte, int, error) {
	return []byte(TARE_CMD_T2200), CMD_TIMEOUT_IMMEDIATE, nil
}

func readWeightCmdT2200() ([]byte, int, error) {
	return []byte(READ_WEIGHT_CMD_T2200), CMD_TIMEOUT_SHORT_100_MS, nil
}

func rebootCmdT2200() ([]byte, int, error) {
	return REBOOT_CMD_T2200, CMD_TIMEOUT_SHORT_100_MS, nil
}

func parseWrDataT2200(inData string) (addr int64, data []byte) { // inData is hex ascii
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

// 打开工厂模式
func enFacModeCmdT2200() ([]byte, int, error) {
	// 构建包头
	packet := make([]byte, OPEN_FAC_CHUNK_SIZE)
	binary.BigEndian.PutUint16(packet[0:2], PACKET_HEAD)

	// 构建命令ID与命令类型
	packet[2] = 0x05
	packet[3] = 0xF1

	// 添加包尾
	binary.BigEndian.PutUint16(packet[OPEN_FAC_CHUNK_SIZE-2:], PACKET_TAIL)

	return packet, CMD_TIMEOUT_SHORT_100_MS, nil
}

// 关闭工厂模式
func disFacModeCmdT2200() ([]byte, int, error) {
	// 构建包头
	packet := make([]byte, OPEN_FAC_CHUNK_SIZE)
	binary.BigEndian.PutUint16(packet[0:2], PACKET_HEAD)

	// 构建命令ID与命令类型
	packet[2] = 0x05
	packet[3] = 0xF2

	// 添加包尾
	binary.BigEndian.PutUint16(packet[OPEN_FAC_CHUNK_SIZE-2:], PACKET_TAIL)

	return packet, CMD_TIMEOUT_SHORT_100_MS, nil
}

// 构建一个数据包
func wrDataCmdT2200(addr uint32, data []byte) ([]byte, int, error) {
	// 构建包头
	packet := make([]byte, FILE_CHUNK_SIZE)
	binary.BigEndian.PutUint16(packet[0:2], PACKET_HEAD)

	// 构建命令ID与命令类型
	packet[2] = CMD_IDENTIFY
	packet[3] = CMD_FLASH

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
	fmt.Printf("%x\n", packet)
	return packet, CMD_TIMEOUT_MEDIUM_500_MS, nil
}

// 擦除原本秤上的打印格式
func eraseCmdT2200(addr uint32) ([]byte, int, error) { // erase size will 2K
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

	fmt.Printf("%x\n", packet)
	return packet, CMD_TIMEOUT_LONG_5000_MS, nil
}
