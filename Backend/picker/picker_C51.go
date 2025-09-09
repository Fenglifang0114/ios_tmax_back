package picker

import (
	"tmaxsrv/comm"
)

const (
	C51_MSG_MAX_LEN = 2048
	C51_MSG_MIN_LEN = 4 // head:2, len: 2, no_data: 1, crc: 4, tail: 2
	// Scale response message and the length of fields
	C51_HEAD_SIZE     = 2
	C51_DATA_LEN_SIZE = 2 // not include head, data length only
	C51_CRC_SIZE      = 4
	C51_TAIL_SIZE     = 2
)

func PickerFnC51(inData []byte, dataLen int) (packOffset uint, packLen uint, shouldRemoveLen uint, pack comm.Packet) {
	// 检查最小长度
	if len(inData) < C51_MSG_MIN_LEN {
		return 0, 0, 0, comm.Packet{}
	}
	// 检查最大长度
	if len(inData) > C51_MSG_MAX_LEN {
		return 0, 0, uint(dataLen), comm.Packet{}
	}

	// 使用现有的findLfCrPos函数查找0d0a位置
	lfcrPos := findLfCrPos(inData, 0, dataLen)

	// 如果找到了0d0a
	if lfcrPos > 0 {
		// 检查数据包总长度是否满足最小长度要求（包括0d0a）
		if lfcrPos+2 < 4 {
			// 如果总长度不足4字节，移除这部分数据并继续等待
			return 0, 0, uint(lfcrPos + 2), comm.Packet{}
		}

		// 构造数据包
		packet := comm.Packet{
			PayloadLen: uint16(lfcrPos),
			CmdID:      0,
			CmdSubId:   0,
			SeqNum:     0,
			Payload:    inData[:lfcrPos],
		}

		packOffset = 0
		packLen = uint(lfcrPos + 2)         // 包括0d0a
		shouldRemoveLen = uint(lfcrPos + 2) // 包括0d0a
		pack = packet
		return
	}

	// 没有找到完整的数据包，等待更多数据
	return 0, 0, 0, comm.Packet{}
}
