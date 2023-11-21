package picker

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"tmaxsrv/comm"
	"tmaxsrv/log"
	"tmaxsrv/util"
)

const (
	CMD_HEAD1 byte = 0x5A
	CMD_HEAD2 byte = 0xA5
	CMD_TAIL1 byte = 0xA5
	CMD_TAIL2 byte = 0x5A

	HEADER_LEN int = 2

	CMD_CONF_TYPE           byte = 0x80
	CMD_CAL0_TYPE           byte = 0x81
	CMD_CALN_TYPE           byte = 0x82
	CMD_LOCK_TYPE           byte = 0x83
	CMD_ZERO_TYPE           byte = 0x84
	CMD_TARE_TYPE           byte = 0x85
	CMD_R1_TYPE             byte = 0x86
	CMD_R1_CONT_TYPE        byte = 0x87
	CMD_R10_TYPE            byte = 0x89
	CMD_R10_CONT_TYPE       byte = 0x8A
	CMD_CODE_TYPE           byte = 0x8C
	CMD_CODE_CONT_TYPE      byte = 0x8D
	CMD_CODE_CONT_STOP_TYPE byte = 0x8E
	RECV_CODE_TYPE          byte = 0xA2

	CMD_PACK_NO_DATA_LEN = 8

	CMD_LEN_IDX = 2
	CMD_SEQ_IDX = 5

	DUMMY_SEQ byte = 0
)

func pickerFnOldScale(inData []byte, dataLen int) (packOffset uint, packLen uint, shouldRemoveLen uint, pack comm.Packet) {
	lfcrPos := findLfCrPos(inData, 0, dataLen)
	if lfcrPos > 0 {
		return 0, uint(lfcrPos), uint(lfcrPos) + 2, comm.Packet{}
	} else {
		return 0, 0, 0, comm.Packet{}
	}
}

// var RESP_OK_DATA = []byte{0x5a, 0xa5, 0x00, 0x01, 0x06, 0x7f, 0xfb, 0xf1, 0x6e, 0xa5, 0x5a}
// var RESP_NAK_DATA = []byte{0x5a, 0xa5, 0x00, 0x01, 0x15, 0x72, 0xbb, 0xd7, 0xb7, 0xa5, 0x5a}
// var RESP_SERIAL_ERROR = []byte{0x5a, 0xa5, 0x00, 0x01, 0x7f, 0xBD, 0x86, 0x1C, 0x86, 0xa5, 0x5a}

const ( // Scale response message and the length of fields
	HEAD_SIZE       int = 2
	DATA_LEN_SIZE       = 2 // not include head
	CMD_ID_SIZE         = 1
	CMD_SUB_ID_SIZE     = 1
	SEQ_NO_SIZE         = 1
	CRC_SIZE            = 4
	TAIL_SIZE           = 2
)

// // Packet 是解析后的数据结构体
// type Packet struct {
// 	PayloadLen uint16
// 	CmdID      uint8
// 	CmdSubId   uint8
// 	SeqNum     uint8
// 	Payload    []byte
// }

const (
	TMAX_MSG_MIN_LEN = 13   // head:2, len: 2, mcd: 1, subtype: 1, seqno: 1, nodata: 0, crc: 4, tail: 2
	TMAX_MSG_MAX_LEN = 2048 // TODO: check if this is correct
)

func pickerFnTmaxPassth(inData []byte, dataLen int) (packOffset uint, packLen uint, shouldRemoveLen uint, pack comm.Packet) {
	packOffset = 0
	packLen = uint(dataLen)
	shouldRemoveLen = uint(dataLen)
	pack.CmdID = 0xff
	pack.CmdSubId = 0x55
	pack.PayloadLen = uint16(dataLen)
	pack.Payload = inData
	return
}

func pickerFnTmax(inData []byte, dataLen int) (packOffset uint, packLen uint, shouldRemoveLen uint, pack comm.Packet) {
	if len(inData) < TMAX_MSG_MIN_LEN {
		return 0, 0, 0, comm.Packet{}
	}

	if len(inData) > TMAX_MSG_MAX_LEN {
		return 0, 0, uint(dataLen), comm.Packet{}
	}
	lastHeadPos := -1
	curpos := 0
	for {
		// find header
		headPos := findHeadPos(inData, curpos, dataLen)
		if headPos == -1 && lastHeadPos == -1 { // no head ever found
			return 0, 0, uint(dataLen), comm.Packet{}
		}
		if headPos != -1 {
			lastHeadPos = headPos // save the last found HeadPos
		}
		if headPos == -1 && lastHeadPos != -1 {
			return 0, 0, uint(headPos + 1), comm.Packet{} // remove data before header
		}
		// check tail
		tailPos, state := verifyTail(inData, headPos, dataLen)
		if state == NOT_ENOUGH_DATA {
			return 0, 0, 0, comm.Packet{}
		} else if state == NOT_FOUND {
			curpos += 2 // skip current header
			continue
		}
		if !verifyCrc(inData, headPos, tailPos) {
			return 0, 0, uint(tailPos + TAIL_SIZE), comm.Packet{}
		}

		// 解析数据
		packet := comm.Packet{
			PayloadLen: binary.BigEndian.Uint16(inData[headPos+2:headPos+2+2]) - DATA_LEN_SIZE - CMD_ID_SIZE - CMD_SUB_ID_SIZE - SEQ_NO_SIZE - CRC_SIZE - TAIL_SIZE,
			CmdID:      inData[headPos+4],
			CmdSubId:   inData[headPos+5],
			SeqNum:     inData[headPos+6],
			Payload:    inData[headPos+7 : tailPos-CRC_SIZE],
		}
		packOffset = uint(headPos)
		packLen = uint(tailPos - headPos + TAIL_SIZE)
		shouldRemoveLen = uint(tailPos + TAIL_SIZE)
		pack = packet
		return
	}
}

func findLfCrPos(buf []byte, offset int, len int) int {
	foundPos := -1

	if len <= 2 { // \r\n
		return -1
	}

	for pos := offset; pos < len-1; pos++ {
		if buf[pos] == 0x0d && buf[pos+1] == 0x0a { // find a correct response
			foundPos = pos
			break
		}
	}

	return foundPos
}

func pickerFnAutoWeight(inData []byte, dataLen int) (packOffset uint, packLen uint, shouldRemoveLen uint) {
	// find and skip header
	headPos := findHeadPos(inData, 0, len(inData))
	if headPos < 0 {
		return 0, 0, 0
	}
	tmpPackLen := inData[headPos+2]
	if int(tmpPackLen) <= dataLen-2 {
		// check if tail and checksum is correct
		if inData[headPos+int(tmpPackLen)] != 0xA5 || inData[headPos+int(tmpPackLen)+1] != 0x5A {
			return 0, 0, uint(dataLen) // packet error
		}

		if inData[headPos+int(tmpPackLen)-4] != GetChkSum(inData[headPos+2 : headPos+int(tmpPackLen)-4])[0] { // -6: checksum(4 bytes) and tail(2 bytes)
			return 0, 0, uint(dataLen) // remove all data in case checksum error
		}

		if tmpPackLen+2 > 18 {
			log.Log.Warn("Error: remove too much data")
		}
		return uint(headPos), uint(tmpPackLen - 7), uint(tmpPackLen) + 2 // not include cmdid(1), checksum(4) and tail(2)
	}

	return 0, 0, 0
}

func findHeadPos(buf []byte, offset int, len int) int {
	foundPos := bytes.Index(buf[offset:], []byte{CMD_HEAD1, CMD_HEAD2})
	if foundPos == -1 {
		return -1
	}

	return foundPos + offset
}

func verifyTail(buf []byte, headPos int, len int) (int, State) {
	if len < headPos+4 { // to prevent index out of bounds
		return -1, NOT_ENOUGH_DATA
	}

	end := headPos + int(binary.BigEndian.Uint16(buf[headPos+2:headPos+2+2])) + HEAD_SIZE
	if end > len { // not enough data entering
		return -1, NOT_ENOUGH_DATA
	}
	if !bytes.Equal(buf[end-TAIL_SIZE:end], []byte{CMD_TAIL1, CMD_TAIL2}) {
		fmt.Println("Invalid footer")
		for _, b := range buf { // print the data
			fmt.Printf("%02x ", b)
		}
		return -1, NOT_FOUND
	}
	return end - TAIL_SIZE, FOUND
}

func verifyCrc(buf []byte, headPos int, tailPos int) bool {
	crc := binary.BigEndian.Uint32(buf[tailPos-CRC_SIZE : tailPos])
	if crc != util.Crc32MPEG2(buf[headPos+HEAD_SIZE:tailPos-CRC_SIZE]) {
		fmt.Println("Invalid CRC")
		return false
	}

	return true
}

func extractPack(b []byte) ([]byte, bool) {
	// check if tail and checksum is correct
	if b[len(b)-2] != 0xA5 || b[len(b)-1] != 0x5A {
		return nil, false
	}

	if b[len(b)-6] != GetChkSum(b[:len(b)-6])[0] { // -6: checksum(4 bytes) and tail(2 bytes)
		return nil, false
	}

	return b[2:(len(b) - 6)], true // remove header, checksum and tail
}

func GetChkSum(s []byte) []byte {
	sumArr := make([]byte, 4)
	var sum byte = 0x0
	for _, v := range s {
		sum = sum ^ byte(v)
	}

	sumArr[0] = sum
	sumArr[1] = 0x0
	sumArr[2] = 0x0
	sumArr[3] = 0x0

	return sumArr
}
