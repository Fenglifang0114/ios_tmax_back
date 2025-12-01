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
	Lp50_INIT      = "FK\"AA\"\r\nFS\"AA\"\r\nD4\r\n"
	LP50_LINE_HEAD = "A"
	Lp50_TEXT_SEP  = "\"" // 文本分割符""
	Lp50_LINE_END  = "\r\n"
	Lp50_LINE_NUM  = 0
	Lp50_TAIL      = "FE\r\nFR\"AA\"\r\nP1,1\r\n"
	Lp50_LOOP_FLAG = 0
	Lp50_FONT_STR  = ",0,4,1,1,N,\"" //票据暂时设置都是这么大小的字体

	//12 23 0F 打印机浓度  00-0F
)

const (
	LP50InnerLineSep    = ","
	Lp50_BAR_CODE_EXCEL = "barcode.xlsx"
)

var FirstLoopY = 0  // 记录第一行循环变量的Y坐标
var SecondLoopY = 0 // 记录第二行循环变量的Y坐标
var LoopYNum = 0    // 记录循环变量的Y行数数量 ，单片机通过这个值来判断要偏移Y的坐标，偏移时要   行数*30

func ParseLP50Lines(buff string, dataBuffer *bytes.Buffer, lastvarPos int) *bytes.Buffer {
	currentPath := comm.GetSrvDataPath()                         //20241107
	VarTable = ReadTableFromFile(currentPath + "/varTable.json") // 获取变量ID表
	// 处理字符串并解析
	buf := dataBuffer
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(buff))

	for scanner.Scan() {
		lines = append(lines, scanner.Text()) // 每行字符串添加到切片
	}

	if len(lines) == 0 {
		return buf
	}
	var countLoop = 0
	var firstLoopIndex = 0
	var secondLoopIndex = 0
	FirstLoopY = 0
	SecondLoopY = 0
	LoopYNum = 0
	//去掉负数位置
	for i := 0; i < len(lines); i++ {
		rowArray := strings.Split(lines[i], ",")
		if strings.Contains(lines[i], "TB") && len(rowArray) > 10 {
			rowArray[1] = strings.ReplaceAll(rowArray[1], "-", "")
			rowArray[2] = strings.ReplaceAll(rowArray[2], "-", "")
			lines[i] = strings.Join(rowArray, ",")
		}
	}

	//"TB,0,250,0,50,4,1,1,0,0,DATA,StartLoop,,1,0,21"
	for i := 0; i < len(lines); i++ {
		rowArray := strings.Split(lines[i], ",")
		if strings.Contains(lines[i], "DATA,StartLoop") {
			countLoop += 1
			if countLoop == 1 && len(rowArray) > 2 {
				firstLoopIndex = i
				FirstLoopY, _ = strconv.Atoi(rowArray[2])
			} else if countLoop == 2 && len(rowArray) > 2 {
				secondLoopIndex = i
				SecondLoopY, _ = strconv.Atoi(rowArray[2])
				break
			}
		}
	}

	LoopYNum = int(math.Round(math.Abs(float64(SecondLoopY-FirstLoopY) / 31)))

	if LoopYNum > 0 && strings.Contains(lines[firstLoopIndex], "DATA,StartLoop") && strings.Contains(lines[secondLoopIndex], "DATA,StartLoop") {
		firstLine := lines[firstLoopIndex]
		parts := strings.Split(firstLine, ",")
		parts[14] = strconv.Itoa(LoopYNum)
		result := strings.Join(parts, ",")
		lines[firstLoopIndex] = result
		secondLine := lines[secondLoopIndex]
		parts2 := strings.Split(secondLine, ",")
		parts2[14] = strconv.Itoa(LoopYNum)
		result2 := strings.Join(parts2, ",")
		lines[secondLoopIndex] = result2

		//处理Y坐标
		for i := 0; i < len(lines); i++ {
			rowArray := strings.Split(lines[i], ",")
			if len(rowArray) > 10 {
				var yPos, _ = strconv.Atoi(rowArray[2])
				if yPos >= FirstLoopY && i != firstLoopIndex && i != secondLoopIndex {
					rowArray[2] = "*#" + rowArray[2]
					lines[i] = strings.Join(rowArray, ",")
				}
			}
		}
	}

	for i := 0; i < len(lines); i++ {
		rowArray := strings.Split(lines[i], ",")
		buf, lastvarPos = Lp50Lines(rowArray, buf, lastvarPos, currentPath)
	}

	fmt.Println("Lp50Lines lastvarPos:", buf)

	// 写入打印命令结尾
	buf.Write([]byte(Lp50_TAIL))
	return buf
}

func Lp50Lines(line []string, dataBuffer *bytes.Buffer, lastvarPos int, path string) (*bytes.Buffer, int) {
	switch line[0] {
	case "P":
		ESC_LINE_NUM = 0
		ESC_LOOP_FLAG = 0
		ParseLp50Page(line, dataBuffer)
	case "TB":
		// fmt.Println(len(line))
		if line[10] == "TEXT" && len(line) == 13 {
			dataBuffer = ParseLp50Text(line, dataBuffer)
		} else if line[10] == "DATA" && len(line) == 16 {
			dataBuffer = ParseLp50Var(line, dataBuffer)
		}
	}
	return dataBuffer, lastvarPos
}

func ParseLp50Page(tempRowArr []string, dataBuffer *bytes.Buffer) *bytes.Buffer {
	dataBuffer.Write([]byte(Lp50_INIT))
	return dataBuffer
}

// TB,12,43,130,30,0,1,1,2,0,TEXT,TIME,0
func ParseLp50Text(tempRowArr []string, dataBuffer *bytes.Buffer) *bytes.Buffer {
	dataBuffer = Lp50TextVarPosInfo(tempRowArr, dataBuffer)
	if len(tempRowArr[11]) > 0 {
		dataBuffer.WriteString(tempRowArr[11])
	}
	dataBuffer.WriteString(Lp50_TEXT_SEP)
	dataBuffer.WriteString(Lp50_LINE_END)

	return dataBuffer
}

func ParseLp50Var(tempRowArr []string, dataBuffer *bytes.Buffer) *bytes.Buffer {
	var tempVarData RptVarStruct
	varId, isFind := findLp50VarID(tempRowArr[11])
	if varId == 254 && ESC_LOOP_FLAG == 0 {
		ESC_LOOP_FLAG++
	} else if varId == 254 && ESC_LOOP_FLAG == 1 {
		varId = 255
		ESC_LOOP_FLAG = 0
	}
	if isFind {
		dataBuffer = Lp50TextVarPosInfo(tempRowArr, dataBuffer)

		// 变量为日期格式需要单独处理
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
			dataBuffer = VarEscDateTime(dataBuffer, varId)
		}

		dataBuffer.WriteString(Lp50_TEXT_SEP)
		dataBuffer.WriteString(Lp50_LINE_END)

	} else {
		fmt.Println("can't find var id ,please check var name")
	}

	return dataBuffer
}

func insertLp50RowEnd(tempYLine int) []byte {
	count := tempYLine - ESC_LINE_NUM
	escLineEnd := make([]byte, count*len(ESC_LINE_END)) // 为了性能考虑，避免在循环中多次调用 Write
	for i := 0; i < count; i++ {
		copy(escLineEnd[i*len(ESC_LINE_END):], ESC_LINE_END)
	}

	ESC_LINE_NUM += count

	return escLineEnd
}

func Lp50TextVarPosInfo(tempRowArr []string, dataBuffer *bytes.Buffer) *bytes.Buffer {

	dataBuffer.WriteString(LP50_LINE_HEAD)
	dataBuffer.WriteString(tempRowArr[1])
	dataBuffer.WriteString(LP50InnerLineSep)
	dataBuffer.WriteString(tempRowArr[2])
	dataBuffer.WriteString(Lp50_FONT_STR)

	return dataBuffer
}

func findLp50VarID(name string) (int, bool) {
	lenth := len(VarTable.ScaleVarTable)
	isFind := false
	id := 0
	for i := 0; i < lenth; i++ {
		if name == VarTable.ScaleVarTable[i].ValueVame {
			id = VarTable.ScaleVarTable[i].Id
			isFind = true
			i = lenth
		}
	}
	return id, isFind
}

func VarLp50DateTime(dataBuffer *bytes.Buffer, varId int) *bytes.Buffer {
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
			tempVarData.maxLen = uint8(2)
			tempVarData.id = uint8(i)
			tempVarData.startPos = uint16(dataBuffer.Len())
			tempVarData.align = uint8(1)
			RptVarList = append(RptVarList, tempVarData)
		}
	}
	return dataBuffer
}

func FindLp50VarInfo(parseContentArr []string, dataBuffer *bytes.Buffer, lastvarPos int) (*bytes.Buffer, int) {
	var tempVarData VarStruct

	for i := 0; i < len(parseContentArr)-1; i++ {
		if parseContentArr[i] == "TEXT" {
			dataBuffer.WriteString(parseContentArr[i+1])
			i = i + 1
		} else if parseContentArr[i] == "DATA" {
			varId, isFind := findLp50VarID(parseContentArr[i+1])
			i = i + 1
			if isFind {
				if varId != 1 && varId != 2 {

					tempVarData.id = uint16(varId)
					tempVarData.startPos = uint16(lastvarPos)
					tempVarData.endPos = uint16(dataBuffer.Len() - lastvarPos)

					i = i + 1
					alignData, err := strconv.Atoi(parseContentArr[i+1])
					if err != nil {
						fmt.Println("data alignment error,please check it")
					}
					tempVarData.align = uint16(alignData)
					i = i + 1
					maxLen, err := strconv.Atoi(parseContentArr[i+1])
					if err != nil {
						fmt.Println("data maxLength error,please check it")
					}
					i = i + 1
					tempVarData.maxlen = uint16(maxLen)
					VarList = append(VarList, tempVarData)
					lastvarPos = int(tempVarData.endPos) + lastvarPos

				} else {
					// dataBuffer = VarEscDateTime(dataBuffer, varId)
				}
			} else {
				fmt.Println("Var name is not find ")
			}
		}
	}
	return dataBuffer, lastvarPos
}
