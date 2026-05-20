package picker

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"tmaxsrv/comm"
	"tmaxsrv/util"
)

const (
	T2200_MSG_MAX_LEN = 2048
	T2200_MSG_MIN_LEN = 11 // head:2, len: 2, no_data: 1, crc: 4, tail: 2
	// Scale response message and the length of fields
	T2200_HEAD_SIZE     = 2
	T2200_DATA_LEN_SIZE = 2 // not include head, data length only
	T2200_CRC_SIZE      = 4
	T2200_TAIL_SIZE     = 2
)

var RESP_OK_DATA = []byte{0x5a, 0xa5, 0x00, 0x01, 0x06, 0x7f, 0xfb, 0xf1, 0x6e, 0xa5, 0x5a}
var RESP_NAK_DATA = []byte{0x5a, 0xa5, 0x00, 0x01, 0x15, 0x72, 0xbb, 0xd7, 0xb7, 0xa5, 0x5a}
var RESP_SERIAL_ERROR = []byte{0x5a, 0xa5, 0x00, 0x01, 0x7f, 0xBD, 0x86, 0x1C, 0x86, 0xa5, 0x5a}

func PickerFnT2200(inData []byte, dataLen int) (packOffset uint, packLen uint, shouldRemoveLen uint, pack comm.Packet) {
	if len(inData) < len(RESP_OK_DATA) {
		return 0, 0, 0, comm.Packet{}
	}

	if len(inData) > T2200_MSG_MAX_LEN {
		return 0, 0, uint(dataLen), comm.Packet{}
	}
	lastHeadPos := -1
	curpos := 0
	for {
		// find header
		headPos := findHeadPos(inData, curpos, dataLen)
		if headPos == -1 && lastHeadPos == -1 { // no head ever found
			// try to find weight data
			lfcrPos := findLfCrPos(inData, 0, dataLen)
			if lfcrPos > 0 {
				return 0, uint(lfcrPos), uint(lfcrPos + 2), comm.Packet{PayloadLen: uint16(lfcrPos), CmdID: 0, CmdSubId: 0, SeqNum: 0, Payload: inData[:lfcrPos]}
			}
			return 0, 0, uint(dataLen), comm.Packet{}
		}
		if headPos != -1 {
			lastHeadPos = headPos // save the last found HeadPos
		}
		if headPos == -1 && lastHeadPos != -1 {
			return 0, 0, uint(lastHeadPos + 2), comm.Packet{} // remove data before header
		}
		// check tail
		tailPos, state := verifyTailT2200(inData, headPos, dataLen)
		if state == NOT_ENOUGH_DATA {
			return 0, 0, 0, comm.Packet{}
		} else if state == NOT_FOUND {
			curpos += 2 // skip current header
			continue
		}
		if !verifyCrcT2200(inData, headPos, tailPos) {
			return 0, 0, uint(tailPos + T2200_TAIL_SIZE), comm.Packet{}
		}

		// 解析数据
		packet := comm.Packet{
			PayloadLen: binary.BigEndian.Uint16(inData[headPos+2 : headPos+2+2]),
			Payload:    inData[headPos+4 : tailPos-T2200_CRC_SIZE],
		}
		packOffset = uint(headPos)
		packLen = uint(tailPos - headPos + T2200_TAIL_SIZE)
		shouldRemoveLen = uint(tailPos + T2200_TAIL_SIZE)
		pack = packet
		return
	}
}

func verifyTailT2200(buf []byte, headPos int, len int) (int, State) {
	if len < headPos+4 { // to prevent index out of bounds
		return -1, NOT_ENOUGH_DATA
	}

	end := headPos + T2200_HEAD_SIZE + T2200_DATA_LEN_SIZE + int(binary.BigEndian.Uint16(buf[headPos+2:headPos+2+2])) + T2200_CRC_SIZE + T2200_TAIL_SIZE
	if end > len { // not enough data entering
		return -1, NOT_ENOUGH_DATA
	}
	if !bytes.Equal(buf[end-T2200_TAIL_SIZE:end], []byte{CMD_TAIL1, CMD_TAIL2}) {
		fmt.Println("Invalid footer")
		for _, b := range buf { // print the data
			fmt.Printf("%02x ", b)
		}
		return -1, NOT_FOUND
	}
	return end - T2200_TAIL_SIZE, FOUND
}

func verifyCrcT2200(buf []byte, headPos int, tailPos int) bool {
	if len(buf) < tailPos+2 || len(buf) < headPos+T2200_MSG_MIN_LEN {
		return false
	}

	crc := binary.BigEndian.Uint32(buf[tailPos-T2200_CRC_SIZE : tailPos])
	if crc != util.Crc32MPEG2(buf[headPos+T2200_HEAD_SIZE:tailPos-T2200_CRC_SIZE]) {
		fmt.Println("Invalid CRC")
		return false
	}

	return true
}
