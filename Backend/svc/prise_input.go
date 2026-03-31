package svc

import (
	"fmt"
	"log"
	"time"
)

// Modbus RTU 地址和功能码
const (
	slaveID   = 0x01
	fcReadInp = 0x02 // 读取输入状态
)

// 请求帧：读取 4 个输入位（起始地址 0，数量 4）
var requestFrame = []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x04}

var buffer []byte // 全局或函数内维持的缓冲区

// 计算 CRC16 (Modbus RTU)
func crc16(data []byte) uint16 {
	var crc uint16 = 0xFFFF
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&0x0001 != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}

// 将 CRC 附加到帧末尾（小端序，低字节在前）
func appendCRC(frame []byte) []byte {
	crc := crc16(frame)
	frame = append(frame, byte(crc&0xFF), byte(crc>>8))
	return frame
}

// 验证接收帧的 CRC，并返回有效数据部分
func verifyCRC(frame []byte) ([]byte, error) {
	fmt.Printf("% X\n", frame)
	if len(frame) < 2 {
		return nil, fmt.Errorf("帧太短")
	}
	recvCRC := uint16(frame[len(frame)-2]) | (uint16(frame[len(frame)-1]) << 8)
	calcCRC := crc16(frame[:len(frame)-2])
	if recvCRC != calcCRC {
		return nil, fmt.Errorf("CRC 校验失败，收到 %04X，计算 %04X", recvCRC, calcCRC)
	}
	return frame[:len(frame)-2], nil
}

// 解析响应，返回数据部分（去掉了地址、功能码、长度）
func parseResponse(response []byte) ([]byte, error) {
	if len(response) < 3 {
		return nil, fmt.Errorf("响应过短")
	}
	if response[0] != slaveID {
		return nil, fmt.Errorf("从站地址不匹配")
	}
	if response[1] != fcReadInp {
		return nil, fmt.Errorf("功能码不匹配")
	}
	dataLen := int(response[2])
	if len(response) < 3+dataLen {
		return nil, fmt.Errorf("数据长度不足")
	}
	return response[3 : 3+dataLen], nil
}

// 将数据字节转换为 4 个开关状态（位 0~3）
func getSwitchStates(data []byte) (states [4]bool) {
	if len(data) == 0 {
		return
	}
	val := data[0]
	states[0] = (val & 0x01) != 0 // 开关1
	states[1] = (val & 0x02) != 0 // 开关2
	states[2] = (val & 0x04) != 0 // 开关3
	states[3] = (val & 0x08) != 0 // 开关4
	return
}

func (sm *SrvMgr) parseFrames() {
	sm.parseMu.Lock()
	defer sm.parseMu.Unlock()

	// 只要缓冲区长度 >= 6，就尝试解析
	for len(sm.parseBuffer) >= 6 {
		candidate := sm.parseBuffer[:6] // 取前 6 字节

		// 校验 CRC
		validFrame, err := verifyCRC(candidate)
		if err == nil {
			// 校验通过，从缓冲区中移除这 6 字节
			sm.parseBuffer = sm.parseBuffer[6:]

			// 解析响应数据
			data, err := parseResponse(validFrame)
			if err != nil {
				log.Println("解析响应错误:", err)
				continue
			}

			// 获取当前开关状态
			currentStates := getSwitchStates(data)
			fmt.Printf("当前状态: %v\n", currentStates)

			// 计算按下的开关数量
			pressedCount := 0
			for i := 0; i < 4; i++ {
				if currentStates[i] {
					pressedCount++
				}
			}

			// 只有 0 或 1 个开关按下时才处理事件
			if pressedCount <= 1 {
				for i := 0; i < 4; i++ {
					if sm.lastStates[i] && !currentStates[i] {
						// fmt.Printf("开关 %d 被按下\n", i+1)
						sm.recvScaleMgrMsg <- &ScaleMgrRespMsg{MsgType: SCALE_MGR_RESP_SCALE_INPUT, MsgBody: fmt.Sprintf("%d", i+1)}
					}
				}
				sm.lastStates = currentStates
			}
		} else {
			// CRC 校验失败，丢弃第一个字节（滑动窗口）
			sm.parseBuffer = sm.parseBuffer[1:]
		}

		time.Sleep(10 * time.Millisecond)
	}
}
