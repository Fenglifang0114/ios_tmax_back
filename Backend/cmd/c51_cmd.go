package cmd

import (
	"encoding/binary"
	"fmt"

	m "tmaxsrv/comm"
)

const (
	PACKET_HEAD_C51         = 0x5AA5
	CMD_IDENTIFY_C51        = 0xA0
	PACKET_TAIL_C51         = 0xA55A
	FILE_CHUNK_SIZE_C51     = 272
	OPEN_FAC_CHUNK_SIZE_C51 = 6
	REC_CHUNK_SIZE_C51      = 11
	EARSE_CHUNK_SIZE_C51    = 0x10
	CMD_ERASE_C51           = 0xA2
	ERASE_SIZE_C51          = 0x800
	CMD_FLASH_C51           = 0xB0
)

const (
	READ_WEIGHT_CMD_C51 string = "W"
	TARE_CMD_C51        string = "T"
	ZERO_CMD_C51        string = "Z"
)

func NewComposerC51() *m.CmdComposer {

	composer := m.CmdComposer{}
	composer.ScaleCat = m.SCALE_C51
	composer.ComposeCmd = ComposeCmdC51

	return &composer
}

func ComposeCmdC51(composer *m.CmdComposer, cmd m.CmdType, cmdData m.CmdData) ([]byte, int, error) {
	if composer.ScaleCat != m.SCALE_C51 {
		return nil, -1, fmt.Errorf("ScaleCat is not SCALE_C51")
	}

	switch cmd {
	// case m.CMD_CHECK_FAC_MODE:
	// 	return enFacModeCmdC51()
	// case m.CMD_DIS_FAC_MODE:
	// 	return disFacModeCmdC51()
	case m.CMD_ZERO:
		return zeroCmdC51()
	case m.CMD_TARE:
		return tareCmdC51()
	case m.CMD_READ_WEIGHT:
		return readWeightCmdC51()

	}
	return nil, -1, fmt.Errorf("cmd not found")
}

func zeroCmdC51() ([]byte, int, error) {
	return []byte(ZERO_CMD_C51), CMD_TIMEOUT_IMMEDIATE, nil
}

func tareCmdC51() ([]byte, int, error) {
	return []byte(TARE_CMD_C51), CMD_TIMEOUT_IMMEDIATE, nil
}

func readWeightCmdC51() ([]byte, int, error) {
	return []byte(READ_WEIGHT_CMD_C51), CMD_TIMEOUT_SHORT_1500_MS, nil
}

// 打开工厂模式
func enFacModeCmdC51() ([]byte, int, error) {
	// 构建包头
	packet := make([]byte, OPEN_FAC_CHUNK_SIZE_C51)
	binary.BigEndian.PutUint16(packet[0:2], PACKET_HEAD_C51)

	// 构建命令ID与命令类型
	packet[2] = 0x05
	packet[3] = 0xF1

	// 添加包尾
	binary.BigEndian.PutUint16(packet[OPEN_FAC_CHUNK_SIZE_C51-2:], PACKET_TAIL_C51)

	return packet, CMD_TIMEOUT_SHORT_1500_MS, nil
}

// 关闭工厂模式
func disFacModeCmdC51() ([]byte, int, error) {
	// 构建包头
	packet := make([]byte, OPEN_FAC_CHUNK_SIZE_C51)
	binary.BigEndian.PutUint16(packet[0:2], PACKET_HEAD_C51)

	// 构建命令ID与命令类型
	packet[2] = 0x05
	packet[3] = 0xF2

	// 添加包尾
	binary.BigEndian.PutUint16(packet[OPEN_FAC_CHUNK_SIZE_C51-2:], PACKET_TAIL_C51)

	return packet, CMD_TIMEOUT_SHORT_1500_MS, nil
}
