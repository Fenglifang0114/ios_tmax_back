package svc

import (
	"encoding/binary"

	utils "tmaxsrv/utils"
)

const (
	FLASH_ADDR          = 0x08003000
	PACKET_HEAD         = 0x5AA5
	CMD_IDENTIFY        = 0xA0
	CMD_TYPE            = 0xB0
	PACKET_TAIL         = 0xA55A
	DATA_LENGTH         = 256
	FILE_CHUNK_SIZE     = 272
	OPEN_FAC_CHUNK_SIZE = 6
	REC_CHUNK_SIZE      = 11
	EARSE_CHUNK_SIZE    = 0x10
	CMD_ERASE           = 0xA2
	CMD_ERASE_SIZE      = 0x800
	CMD_FLASH           = 0xB0
)

const (
	CMD_IDENTIFY_IDX = 2
	CMD_TYPE_IDX     = 3
)

// 打开工厂模式
func enFacModeCmd() []byte {
	// 构建包头
	packet := make([]byte, OPEN_FAC_CHUNK_SIZE)
	binary.BigEndian.PutUint16(packet[0:2], PACKET_HEAD)

	// 构建命令ID与命令类型
	packet[2] = 0x05
	packet[3] = 0xF1

	// 添加包尾
	binary.BigEndian.PutUint16(packet[OPEN_FAC_CHUNK_SIZE-2:], PACKET_TAIL)

	return packet
}

// 构建一个数据包
func wrDataCmd(addr uint32, data []byte) []byte {
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
	checksum := utils.Crc32MPEG2(packet[2 : FILE_CHUNK_SIZE-6])
	binary.BigEndian.PutUint32(packet[FILE_CHUNK_SIZE-6:], checksum)

	// 添加包尾
	binary.BigEndian.PutUint16(packet[FILE_CHUNK_SIZE-2:], PACKET_TAIL)

	return packet
}

// 擦除原本秤上的打印格式
func eraseCmd(addr uint32) []byte { // erase size will 2K
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
	checksum := utils.Crc32MPEG2(packet[2 : EARSE_CHUNK_SIZE-6])
	binary.BigEndian.PutUint32(packet[EARSE_CHUNK_SIZE-6:], checksum)

	// 添加包尾
	binary.BigEndian.PutUint16(packet[EARSE_CHUNK_SIZE-2:], PACKET_TAIL)

	return packet
}
