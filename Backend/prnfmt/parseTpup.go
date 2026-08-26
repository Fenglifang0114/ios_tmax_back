package prnfmt

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"

	"tmaxsrv/comm"
)

var (
	TPUP_INIT      = []byte{0x1B, 0x40, 0x1C, 0x26, 0x1B, 0x39, 0x01}
	TPUP_LINE_END  = []byte{0x0A}
	TPUP_TAIL      = []byte{0x1B, 0x64, 0x03}
	TPUP_LOOP_FLAG = 0
)

const (
	TPUP_LINE_HEIGHT = 31
)

func ParseTpupLines(buff string, dataBuffer *bytes.Buffer, lastvarPos int) *bytes.Buffer {
	currentPath := comm.GetSrvDataPath()
	VarTable = ReadTableFromFile(currentPath + "/varTable.json")

	buf := dataBuffer
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(buff))

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if len(lines) == 0 {
		return buf
	}

	// Remove negative positions
	for i := 0; i < len(lines); i++ {
		rowArray := strings.Split(lines[i], ",")
		if strings.Contains(lines[i], "TB") && len(rowArray) > 10 {
			rowArray[1] = strings.ReplaceAll(rowArray[1], "-", "")
			rowArray[2] = strings.ReplaceAll(rowArray[2], "-", "")
			lines[i] = strings.Join(rowArray, ",")
		}
	}

	var countLoop = 0
	var firstLoopIndex = 0
	var secondLoopIndex = 0
	firstLoopY := 0
	secondLoopY := 0
	loopYNum := 0

	for i := 0; i < len(lines); i++ {
		rowArray := strings.Split(lines[i], ",")
		if strings.Contains(lines[i], "DATA,StartLoop") {
			countLoop += 1
			if countLoop == 1 && len(rowArray) > 2 {
				firstLoopIndex = i
				firstLoopY, _ = strconv.Atoi(rowArray[2])
			} else if countLoop == 2 && len(rowArray) > 2 {
				secondLoopIndex = i
				secondLoopY, _ = strconv.Atoi(rowArray[2])
				break
			}
		}
	}

	loopYNum = int(math.Round(math.Abs(float64(secondLoopY-firstLoopY) / float64(TPUP_LINE_HEIGHT))))

	if loopYNum > 0 && strings.Contains(lines[firstLoopIndex], "DATA,StartLoop") && strings.Contains(lines[secondLoopIndex], "DATA,StartLoop") {
		firstLine := lines[firstLoopIndex]
		parts := strings.Split(firstLine, ",")
		parts[14] = strconv.Itoa(loopYNum)
		lines[firstLoopIndex] = strings.Join(parts, ",")

		secondLine := lines[secondLoopIndex]
		parts2 := strings.Split(secondLine, ",")
		parts2[14] = strconv.Itoa(loopYNum)
		lines[secondLoopIndex] = strings.Join(parts2, ",")

		for i := 0; i < len(lines); i++ {
			rowArray := strings.Split(lines[i], ",")
			if len(rowArray) > 10 {
				yPos, _ := strconv.Atoi(rowArray[2])
				if yPos >= firstLoopY && i != firstLoopIndex && i != secondLoopIndex {
					rowArray[2] = "*#" + rowArray[2]
					lines[i] = strings.Join(rowArray, ",")
				}
			}
		}
	}

	for i := 0; i < len(lines); i++ {
		rowArray := strings.Split(lines[i], ",")
		buf, lastvarPos = TpupLines(rowArray, buf, lastvarPos, currentPath)
	}

	buf.Write(TPUP_TAIL)
	return buf
}

func TpupLines(line []string, dataBuffer *bytes.Buffer, lastvarPos int, path string) (*bytes.Buffer, int) {
	if len(line) == 0 {
		return dataBuffer, lastvarPos
	}
	switch line[0] {
	case "P":
		TPUP_LOOP_FLAG = 0
		ParseTpupPage(line, dataBuffer)
	case "TB":
		if line[10] == "TEXT" && len(line) >= 13 {
			dataBuffer = ParseTpupText(line, dataBuffer)
		} else if line[10] == "DATA" && len(line) >= 16 {
			dataBuffer = ParseTpupVar(line, dataBuffer)
		}
	}
	return dataBuffer, lastvarPos
}

func ParseTpupPage(tempRowArr []string, dataBuffer *bytes.Buffer) *bytes.Buffer {
	dataBuffer.Write(TPUP_INIT)
	return dataBuffer
}

func ParseTpupText(tempRowArr []string, dataBuffer *bytes.Buffer) *bytes.Buffer {
	dataBuffer = TpupTextVarPosInfo(tempRowArr, dataBuffer)
	if len(tempRowArr[11]) > 0 {
		dataBuffer.WriteString(tempRowArr[11])
	}
	return dataBuffer
}

func ParseTpupVar(tempRowArr []string, dataBuffer *bytes.Buffer) *bytes.Buffer {
	var tempVarData RptVarStruct
	varId, isFind := findTpupVarID(tempRowArr[11])
	if varId == 254 && TPUP_LOOP_FLAG == 0 {
		TPUP_LOOP_FLAG++
	} else if varId == 254 && TPUP_LOOP_FLAG == 1 {
		varId = 255
		TPUP_LOOP_FLAG = 0
	}

	if isFind {
		dataBuffer = TpupTextVarPosInfo(tempRowArr, dataBuffer)

		if varId != 1 && varId != 2 {
			tempVarData.id = uint8(varId)
			tempVarData.startPos = uint16(dataBuffer.Len())
			intAlign, err := strconv.Atoi(tempRowArr[13])
			if err == nil {
				tempVarData.align = uint8(intAlign)
			}
			intMaxLen, err := strconv.Atoi(tempRowArr[14])
			if err == nil {
				tempVarData.maxLen = uint8(intMaxLen)
			}
			RptVarList = append(RptVarList, tempVarData)
		} else {
			dataBuffer = VarTpupDateTime(dataBuffer, varId)
		}
	} else {
		fmt.Println("can't find var id, please check var name")
	}

	return dataBuffer
}

func TpupTextVarPosInfo(tempRowArr []string, dataBuffer *bytes.Buffer) *bytes.Buffer {
	xOffset, err := strconv.Atoi(tempRowArr[1])
	if err != nil {
		return dataBuffer
	}
	yOffset, err := strconv.Atoi(tempRowArr[2])
	if err != nil {
		return dataBuffer
	}

	// ESC/POS position command: ESC $ nL nH
	dataBuffer.Write([]byte{0x1B, 0x24, byte(xOffset & 0xFF), byte((xOffset >> 8) & 0xFF)})
	_ = yOffset

	// Font style & size
	if len(tempRowArr) > 8 {
		if tempRowArr[8] == "0" {
			dataBuffer.Write([]byte{0x1B, 0x45, 0x00, 0x1D, 0x42, 0x00})
		} else if tempRowArr[8] == "1" {
			dataBuffer.Write([]byte{0x1B, 0x45, 0x00, 0x1D, 0x42, 0x01})
		} else if tempRowArr[8] == "2" {
			dataBuffer.Write([]byte{0x1B, 0x45, 0x01, 0x1D, 0x42, 0x00})
		} else if tempRowArr[8] == "3" {
			dataBuffer.Write([]byte{0x1B, 0x45, 0x01, 0x1D, 0x42, 0x01})
		}
	}
	return dataBuffer
}

func findTpupVarID(name string) (int, bool) {
	lenth := len(VarTable.ScaleVarTable)
	isFind := false
	id := 0
	for i := 0; i < lenth; i++ {
		if name == VarTable.ScaleVarTable[i].ValueVame {
			id = VarTable.ScaleVarTable[i].Id
			isFind = true
			break
		}
	}
	return id, isFind
}

func VarTpupDateTime(dataBuffer *bytes.Buffer, varId int) *bytes.Buffer {
	var tempVarData RptVarStruct
	if varId == 1 {
		for i := 1; i < 4; i++ {
			if i == 1 {
				tempVarData.maxLen = uint8(4)
			} else {
				dataBuffer.WriteString("/")
				tempVarData.maxLen = uint8(2)
			}
			tempVarData.id = uint8(i)
			tempVarData.startPos = uint16(dataBuffer.Len())
			tempVarData.align = uint8(1)
			RptVarList = append(RptVarList, tempVarData)
		}
	} else {
		for i := 4; i < 7; i++ {
			if i == 4 {
				tempVarData.maxLen = uint8(2)
			} else {
				dataBuffer.WriteString(":")
				tempVarData.maxLen = uint8(2)
			}
			tempVarData.id = uint8(i)
			tempVarData.startPos = uint16(dataBuffer.Len())
			tempVarData.align = uint8(1)
			RptVarList = append(RptVarList, tempVarData)
		}
	}
	return dataBuffer
}
