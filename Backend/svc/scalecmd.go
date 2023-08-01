package svc

import (
	"encoding/binary"
	"fmt"
	"time"

	l "tmaxsrv/log"
	"tmaxsrv/utils"
)

var EN_ENG_CMD []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0x05, 0xf1, 0x00, 0xf9, 0x16, 0x57, 0x5e, 0xa5, 0x5a}
var DIS_ENG_CMD []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0x05, 0xf2, 0x00, 0x8b, 0xfd, 0x08, 0x8d, 0xa5, 0x5a}
var TMAX_ZERO_CMD []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0xe1, 0x03, 0x00, 0x0c, 0xD6, 0x90, 0x0C, 0xa5, 0x5a}
var TMAX_TARE_CMD []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0xe1, 0x05, 0x00, 0xE9, 0x00, 0x2F, 0xAA, 0xa5, 0x5a}
var READ_WEIGHT_CMD []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0xe1, 0x05, 0x00, 0xac, 0x24, 0x0e, 0x03, 0xa5, 0x5a}
var EN_CONT_MODE_CMD []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0xe1, 0x07, 0x00, 0x49, 0xf2, 0xb1, 0xa5, 0xa5, 0x5a}
var DIS_CONT_MODE_CMD []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0xe1, 0x08, 0x00, 0xf4, 0x75, 0x8c, 0x8d, 0xa5, 0x5a}
var EN_PASSTH_MODE_CMD []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0x05, 0xf3, 0x00, 0x59, 0xE4, 0xC9, 0x51, 0xa5, 0x5a}
var DIS_PASSTH_MODE_CMD []byte = []byte{0x5a, 0xa5, 0x00, 0x0b, 0x05, 0xf4, 0x00, 0x6E, 0x2B, 0xB7, 0x2B, 0xa5, 0x5a}
var GET_AP_LIST_CMD []byte = []byte("AT+CWLAP\r\n")

var GExpectWifiResp string

func (c *Scale) PerfZero() bool {
	if c.isOldC51Scale {
		return perfCmd(c, []byte(ZERO_CMD))
	} else {
		_, err := perfCmdNwaitResult(c, TMAX_ZERO_CMD, ZERO_CMD_RESP)
		return err == nil
	}
}

func (c *Scale) PerfTare() bool {
	if c.isOldC51Scale {
		return perfCmd(c, []byte(TARE_CMD))
	} else {
		_, err := perfCmdNwaitResult(c, TMAX_TARE_CMD, TARE_CMD_RESP)
		return err == nil
	}
}

func (c *Scale) ReadWeight() bool {
	if c.isOldC51Scale {
		return perfCmd(c, []byte(GET_WEIGHT_CMD))
	} else {
		_, err := perfCmdNwaitResult(c, READ_WEIGHT_CMD, WEIGHT_DATA_RESP)
		return err == nil
	}
}

func (c *Scale) RegWeightData() bool {
	// enable scale sending weighing info continually
	_, err := perfCmdNwaitResult(c, EN_CONT_MODE_CMD, REG_WEIGHT_RESP)
	c.isSendUnolicitedData = true
	return err == nil
}

func (c *Scale) UnRegWeightData() bool {
	// enable scale sending weighing info continually
	msg, err := perfCmdNwaitResult(c, DIS_CONT_MODE_CMD, UNREG_WEIGHT_RESP)
	msgStr, _ := json.MarshalToString(msg)
	if err != nil {
		if c.client != nil {
			c.client.sendCh <- []byte(msgStr)
		}
	}
	c.isSendUnolicitedData = false
	return err == nil
}

func (c *Scale) GetRecs() ([]ScaleRec, error) {
	return c.scaleMgr.recPb.GetRecsList(*c)
}

func (c *Scale) AddRec(rec ScaleRec) error {
	rec.ScaleModel = c.Model
	rec.ScaleSn = c.Sn
	return c.scaleMgr.recPb.InsertRec(rec)
}

func (c *Scale) DelRec(recId uint) error {
	return c.scaleMgr.recPb.DeleteRec(recId)
}

const (
	FLASH_ADDR                = 0x08003000
	PACKET_HEAD               = 0x5AA5
	CMD_IDENTIFY              = 0xA0
	CMD_TYPE                  = 0xB0
	PACKET_TAIL               = 0xA55A
	DATA_LENGTH               = 256
	FILE_CHUNK_SIZE           = 272
	OPEN_FAC_CHUNK_SIZE       = 6
	CLOSE_FAC_CHUNK_SIZE      = 6
	EN_PASSTH_CHUNK_SIZE      = 6
	DIS_PASSTH_CHUNK_SIZE     = 6
	MODIFY_BT_NAME_CHUNK_SIZE = 24
	REC_CHUNK_SIZE            = 11
	EARSE_CHUNK_SIZE          = 0x10
	CMD_ERASE                 = 0xA2
	CMD_ERASE_SIZE            = 0x800
	CMD_FLASH                 = 0xB0
)

const (
	CMD_IDENTIFY_IDX = 2
	CMD_TYPE_IDX     = 3
)

// // 打开工厂模式命令
// func EnFacModeCmd() []byte {
// 	// 构建包头
// 	packet := make([]byte, OPEN_FAC_CHUNK_SIZE)
// 	binary.BigEndian.PutUint16(packet[0:2], PACKET_HEAD)

// 	// 构建命令ID与命令类型
// 	packet[2] = 0x05
// 	packet[3] = 0xF1

// 	// 添加包尾
// 	binary.BigEndian.PutUint16(packet[OPEN_FAC_CHUNK_SIZE-2:], PACKET_TAIL)

// 	return packet
// }

// // 关闭工厂模式命令
// func DisFacModeCmd() []byte {
// 	// 构建包头
// 	packet := make([]byte, CLOSE_FAC_CHUNK_SIZE)
// 	binary.BigEndian.PutUint16(packet[0:2], PACKET_HEAD)

// 	// 构建命令ID与命令类型
// 	packet[2] = 0x05
// 	packet[3] = 0xF2

// 	// 添加包尾
// 	binary.BigEndian.PutUint16(packet[CLOSE_FAC_CHUNK_SIZE-2:], PACKET_TAIL)

// 	return packet
// }

// func EnPassthCmd() []byte {
// 	// 构建包头
// 	packet := make([]byte, EN_PASSTH_CHUNK_SIZE)
// 	binary.BigEndian.PutUint16(packet[0:2], PACKET_HEAD)

// 	// 构建命令ID与命令类型
// 	packet[2] = uint8(EN_PASSTH_CHUNK_SIZE - HEADER_LEN)
// 	packet[3] = 0x05
// 	packet[4] = 0xF3

// 	// 添加包尾
// 	binary.BigEndian.PutUint16(packet[EN_PASSTH_CHUNK_SIZE-2:], PACKET_TAIL)

// 	return packet
// }

// func DisPassthCmd() []byte {
// 	// 构建包头
// 	packet := make([]byte, DIS_PASSTH_CHUNK_SIZE)
// 	binary.BigEndian.PutUint16(packet[0:2], PACKET_HEAD)

// 	// 构建命令ID与命令类型
// 	packet[2] = uint8(DIS_PASSTH_CHUNK_SIZE - HEADER_LEN)
// 	packet[3] = 0x05
// 	packet[4] = 0xF4

// 	// 添加包尾
// 	binary.BigEndian.PutUint16(packet[DIS_PASSTH_CHUNK_SIZE-2:], PACKET_TAIL)

// 	return packet
// }

// 修改BT名称
func ModifyBTNameCmd(name string) []byte {
	// MODIFY_BT_NAME_CHUNK_SIZE should include all data except BT name
	CMD := "TTM:REN-"
	// 构建包头
	var packLen uint16 = uint16(MODIFY_BT_NAME_CHUNK_SIZE + len(name))
	packet := make([]byte, packLen)
	binary.BigEndian.PutUint16(packet[0:2], PACKET_HEAD)
	// 构建命令ID与命令类型
	binary.BigEndian.PutUint16(packet[2:4], packLen-2)
	var i uint16 = 4
	packet[i] = 0xF2 // packet length without header
	i += 1
	packet[i] = 0x01
	i += 1
	packet[i] = 0x00
	i += 1
	copy(packet[i:], CMD)
	i += uint16(len(CMD))
	copy(packet[i:], name)
	i += uint16(len(name))
	packet[i] = 0x0d
	i += 1
	packet[i] = 0x0a
	i += 1
	packet[i] = 0x00
	i += 1
	if i != packLen-6 {
		panic("packet len error!")
	}
	// 计算与添加校验码
	checksum := utils.Crc32MPEG2(packet[2 : packLen-6])
	binary.BigEndian.PutUint32(packet[packLen-6:], checksum)
	// 添加包尾
	binary.BigEndian.PutUint16(packet[packLen-2:], PACKET_TAIL)

	return packet
}

// 修改BT名称
func ComposeToWifiPassthData(dataStr string) []byte {
	// MODIFY_BT_NAME_CHUNK_SIZE should include all data except BT name
	data := []byte(dataStr)
	// 构建包头
	var packLen uint16 = uint16(13 + len(dataStr)) // 13: Head: 2 + len:2 + CmdId: 1 + SubCmdId: 1 + SeqNo:1 + CRC:4 + Tail:2
	packet := make([]byte, packLen)
	binary.BigEndian.PutUint16(packet[0:2], PACKET_HEAD)
	// 构建命令ID与命令类型
	binary.BigEndian.PutUint16(packet[2:4], packLen-2)
	var i uint16 = 4
	packet[i] = 0xF2 // packet length without header
	i += 1
	packet[i] = 0x02
	i += 1
	packet[i] = 0x00
	i += 1
	copy(packet[i:], data)
	i += uint16(len(data))
	if i != packLen-6 {
		panic("packet len error!")
	}
	// 计算与添加校验码
	checksum := utils.Crc32MPEG2(packet[2 : packLen-6])
	binary.BigEndian.PutUint32(packet[packLen-6:], checksum)
	// 添加包尾
	binary.BigEndian.PutUint16(packet[packLen-2:], PACKET_TAIL)

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

// 打开工厂模式
func EnFacMode(s *Scale) error {
	l.Log.Debug("send enable factory mode cmd to scale")
	cmd := EN_ENG_CMD
	if res, err := perfCmdNwaitResult(s, cmd, EN_FAC_MODE_RESP); err != nil {
		return err
	} else if res.MsgBody != "ok" {
		return fmt.Errorf("enable factory mode fail")
	} else {
		// do nothing
	}

	return nil
}

// 关闭工厂模式
func DisFacMode(s *Scale) error {
	l.Log.Debug("send disable factory mode cmd to scale")
	cmd := DIS_ENG_CMD
	if res, err := perfCmdNwaitResult(s, cmd, DIS_FAC_MODE_RESP); err != nil {
		return err
	} else if res.MsgBody != "ok" {
		return fmt.Errorf("diable factory mode fail")
	} else {
		// do nothing
	}

	return nil
}

// 打开BT透传模式
func EnPassthrough(s *Scale) error {
	l.Log.Debug("send enable passthrough mode cmd to scale")
	cmd := EN_PASSTH_MODE_CMD
	if res, err := perfCmdNwaitResult(s, cmd, EN_PASSTH_MODE_RESP); err != nil {
		return err
	} else if res.MsgBody != "ok" {
		return fmt.Errorf("enable BT passthrough mode fail")
	} else {
		// do nothing
	}

	return nil
}

// 关闭BT透传模式
func DisPassthrough(s *Scale) error {
	l.Log.Debug("send disable BT passthrough mode cmd to scale")
	cmd := DIS_PASSTH_MODE_CMD
	if res, err := perfCmdNwaitResult(s, cmd, DIS_PASSTH_MODE_RESP); err != nil {
		return err
	} else if res.MsgBody != "ok" {
		return fmt.Errorf("disable BT passthrough mode fail")
	} else {
		// do nothing
	}

	return nil
}

// 修改蓝牙名称
func ModifyBTName(s *Scale, name string) error {
	l.Log.Debug("modify BT name")
	cmd := ModifyBTNameCmd(name)
	if res, err := perfCmdNwaitResult(s, cmd, MODIFY_BT_NAME_RESP); err != nil {
		return err
	} else if res.MsgBody != "ok" {
		return fmt.Errorf("modify BT name fail")
	} else {
		// do nothing
	}

	return nil
}

// Get AP list
func GetApList(s *Scale) error {
	l.Log.Debug("Get AP list")
	GExpectWifiResp = "get_ap_list"
	cmd := ComposeToWifiPassthData(string(GET_AP_LIST_CMD))
	if res, err := perfCmdNwaitResult(s, cmd, GET_AP_LIST_RESP); err != nil {
		return err
	} else if res.MsgBody != "ok" {
		return fmt.Errorf("modify BT name fail")
	} else {
		// do nothing
	}

	return nil
}
func perfCmd(c *Scale, cmd []byte) bool {
	if err := writeScale(c, cmd); err != nil {
		l.Log.Error(err.Error())
		return false
	}
	return true
}

func perfCmdNwaitResult(c *Scale, cmd []byte, waitMsgType RespMsgType) (*ScaleRespMsg, error) {
	ch := make(chan *ScaleRespMsg, 10)
	c.RegisterNotif(waitMsgType, ch)
	defer func() {
		c.UnRegisterNotif(waitMsgType, ch)
		c.isWaintingResp = false
	}()

	GlastWantRespMsgType = waitMsgType
	c.isWaintingResp = true
	if err := writeScale(c, cmd); err != nil {
		l.Log.Error(err.Error())
		return &ScaleRespMsg{MsgType: waitMsgType, ScaleId: c.Id, MsgBody: "error"}, err
	}

	var ret *ScaleRespMsg
	select {
	case ret = <-ch:
	case <-time.After(SCALE_TIME_OUT_S * time.Second):
		ret = &ScaleRespMsg{MsgType: waitMsgType, ScaleId: c.Id, MsgBody: "timeout"}
		return ret, fmt.Errorf("no response, time out")
	}

	return ret, nil
}

func writeScale(c *Scale, data []byte) error {
	if c.MySerial == nil {
		return fmt.Errorf("scale without a ComPort")
	}

	if len(c.MySerial.sendCh) > SEND_CH_SIZE {
		return fmt.Errorf("serial sendCh full")
	}
	c.MySerial.sendCh <- data
	return nil
}

// write send messages from the hub to scale.
// A goroutine running write is started for each scale. The
// application ensures that there is at most one writer to a scale by
// executing all writes from this goroutine.
func (c *Scale) write() {
	for {
		select {
		case message, ok := <-c.toScaleMsgCh:
			if !ok {
				// The hub closed the channel.
				return
			}
			// send message to scale
			c.MySerial.Write([]byte(message))
		}
	}
}
