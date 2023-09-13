package prnfmt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"

	"golang.org/x/text/encoding/simplifiedchinese"
)

type InfoStr struct {
	addr      uint32
	formatLen uint32
	varNum    uint32
}
type printInfo struct {
	printerName     [23]byte
	formatNum       uint8
	everyFormatInfo [12]InfoStr
	varInfo         [30]VarStruct
}

type printInfoTmax struct {
	printerName     [23]byte
	formatNum       uint8
	everyFormatInfo [12]InfoStr
	varInfo         [120]VarStruct
}

const (
	headBufLen = 468
)

func ParserFmtToFile(utf8Buff string) bool {
	buffer := ParserFmtToBuf(utf8Buff)

	// 7.创建bin文件

	if creatFile("formatBin.bin", buffer) {
		fmt.Println("creat binary file success")
		return true
	} else {
		fmt.Println("creat binary file fail")
		return false
	}

	// 8.写数据到串口
}

func ParserFmtToBuf(utf8Buff string) *bytes.Buffer {
	var clearList []VarStruct
	VarList = clearList // 用于清空数据

	var clearTable ScaleVarOrder
	VarTable = clearTable // 用于清空数据

	var FinalFormatInfo printInfo
	var everyBufLen []int                    // 每个打印格式的命令集合
	totalbuffer := bytes.NewBufferString("") // 打印命令集合
	dataCamp := bytes.NewBufferString("")
	TotalVarDataIndex := 0 // 每个打印格式信息的索引

	fillchar := 0xff
	lastVarPos := 0
	lastVarNum := 0
	lastAddr := 0

	// buff, _ := Utf8ToGb2312(utf8Buff)
	formatbuf := ParseEplLines(utf8Buff, dataCamp, lastVarPos)
	everyBufLen = append(everyBufLen, formatbuf.Len())

	fmt.Println(string(dataCamp.Bytes()))

	totalbuffer.WriteString(formatbuf.String())
	div := ((everyBufLen[TotalVarDataIndex] / 4) + 1) * 4
	for i := formatbuf.Len(); i < div; i++ {
		totalbuffer.WriteByte(byte(fillchar))
	}

	if TotalVarDataIndex > 0 {
		lastVarNum = len(VarList) - lastVarNum
	} else {
		lastAddr = headBufLen
		lastVarNum = len(VarList)
	}

	FinalFormatInfo.formatNum = 1 ///打印格式总数，根据打印格式文件数量决定
	FinalFormatInfo.everyFormatInfo[TotalVarDataIndex].addr = uint32(lastAddr)
	FinalFormatInfo.everyFormatInfo[TotalVarDataIndex].formatLen = uint32(everyBufLen[TotalVarDataIndex])
	FinalFormatInfo.everyFormatInfo[TotalVarDataIndex].varNum = uint32(lastVarNum)
	dataCamp.Reset()

	TotalVarDataIndex = TotalVarDataIndex + 1
	lastVarNum = len(VarList)
	lastAddr = div + lastAddr

	// 4.将变量信息写入结构体
	for i := 0; i < len(VarList); i++ {
		FinalFormatInfo.varInfo[i] = VarList[i]
	}
	// 5.转换为bin文件
	buffer := binaryData(FinalFormatInfo)
	binary.Write(buffer, binary.LittleEndian, totalbuffer.Bytes())

	if buffer.Len() < (2048 - 8) {
		for i := buffer.Len(); i < (2048 - 8); i++ {
			binary.Write(buffer, binary.LittleEndian, byte(fillchar))
		}
	}

	// 6.写结尾5a a5 a5 5a 00 00 00 00
	fillTail := []byte{0x5a, 0xa5, 0xa5, 0x5a, 0x00, 0x00, 0x00, 0x00}
	dataCamp1 := bytes.NewBufferString("")

	for i := 0; i < len(fillTail); i++ {
		dataCamp1.WriteByte(fillTail[i])
	}

	binary.Write(buffer, binary.LittleEndian, dataCamp1.Bytes())

	return buffer
}

// 转二进制
func binaryData(tempInfo printInfo) *bytes.Buffer {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, tempInfo.printerName)
	binary.Write(buf, binary.LittleEndian, tempInfo.formatNum)
	binary.Write(buf, binary.LittleEndian, tempInfo.everyFormatInfo)
	binary.Write(buf, binary.LittleEndian, tempInfo.varInfo)
	return buf
}

// 创建bin文件
func creatFile(fileName string, tempBuf *bytes.Buffer) bool {
	fp, err := os.Create(fileName)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer fp.Close()
	fp.Write(tempBuf.Bytes())
	return true
}

// Utf8ToGb2312 将UTF-8字符串转换为GB2312编码
func Utf8ToGb2312(buff string) (string, error) {
	utf8str := buff
	enc := simplifiedchinese.GB18030.NewEncoder()
	utf8Bytes := []byte(utf8str)
	gb2312Bytes, err := enc.Bytes(utf8Bytes)
	if err != nil {
		log.Fatal(err)
	}
	gb2312str := string(gb2312Bytes)

	return gb2312str, err
}
