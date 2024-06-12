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

type printInfoRpt struct {
	printerName     [23]byte
	formatNum       uint8
	everyFormatInfo [12]InfoStr
	varInfo         [60]RptVarStruct
}

type printInfoDef struct {
	printerName     [23]byte
	formatNum       uint8
	everyFormatInfo [12]InfoStr    //最大12个打印格式
	varInfo         [150]VarStruct //最大150个变量
}

const (
	headBufLen        = 468  //一个printInfo占用的字节是固定的。
	headBufLenReceipt = 468  //一个printInfoReceipt占用的字节是固定的。
	headBufLenDef     = 1668 //默认打印格式占用的字节是固定的。
)

var (
	FMT_FILL_TAIL []byte = []byte{0x5a, 0xa5, 0xa5, 0x5a, 0x00, 0x00, 0x00, 0x00}
)

func ParserFmtToFile(utf8Buff string, printerModel string, fmtLen int) bool {
	var buffer *bytes.Buffer
	if printerModel == "EPM205" {
		buffer = ParserFmtToBuf(utf8Buff, printerModel, fmtLen)
	} else {
		buffer = ParserRptFmtToBuf(utf8Buff, printerModel, fmtLen)
	}

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

func ParserFmtToBuf(utf8Buff string, printerModel string, fmtLen int) *bytes.Buffer {
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

	buff, _ := Utf8ToGb2312(utf8Buff)
	var formatbuf *bytes.Buffer

	formatbuf = ParseEplLines(buff, dataCamp, lastVarPos)

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

	if buffer.Len() < (fmtLen - 8) {
		for i := buffer.Len(); i < (fmtLen - 8); i++ {
			binary.Write(buffer, binary.LittleEndian, byte(fillchar))
		}
	}

	// 6.写结尾5a a5 a5 5a 00 00 00 00

	dataCamp1 := bytes.NewBufferString("")

	for i := 0; i < len(FMT_FILL_TAIL); i++ {
		dataCamp1.WriteByte(FMT_FILL_TAIL[i])
	}

	binary.Write(buffer, binary.LittleEndian, dataCamp1.Bytes())

	return buffer
}

func ParserRptFmtToBuf(utf8Buff string, printerModel string, fmtLen int) *bytes.Buffer {
	var clearList []RptVarStruct
	RptVarList = clearList // 用于清空数据

	var clearTable ScaleVarOrder
	VarTable = clearTable // 用于清空数据

	var FinalRptFmtInfo printInfoRpt
	var everyBufLen []int                    // 每个打印格式的命令集合
	totalbuffer := bytes.NewBufferString("") // 打印命令集合
	dataCamp := bytes.NewBufferString("")
	TotalVarDataIndex := 0 // 每个打印格式信息的索引

	fillchar := 0xff
	lastVarPos := 0
	lastVarNum := 0
	lastAddr := 0

	// buff, _ := Utf8ToGb2312(utf8Buff)
	buff := utf8Buff //用UTF8 做
	var formatbuf *bytes.Buffer

	formatbuf = ParseEscLines(buff, dataCamp, lastVarPos)

	everyBufLen = append(everyBufLen, formatbuf.Len())

	fmt.Println(string(dataCamp.Bytes()))

	totalbuffer.WriteString(formatbuf.String())
	div := ((everyBufLen[TotalVarDataIndex] / 4) + 1) * 4
	for i := formatbuf.Len(); i < div; i++ {
		totalbuffer.WriteByte(byte(fillchar))
	}

	if TotalVarDataIndex > 0 {
		lastVarNum = len(RptVarList) - lastVarNum
	} else {
		lastAddr = headBufLen
		lastVarNum = len(RptVarList)
	}

	FinalRptFmtInfo.formatNum = 1 ///打印格式总数，根据打印格式文件数量决定
	FinalRptFmtInfo.everyFormatInfo[TotalVarDataIndex].addr = uint32(lastAddr)
	FinalRptFmtInfo.everyFormatInfo[TotalVarDataIndex].formatLen = uint32(everyBufLen[TotalVarDataIndex])
	FinalRptFmtInfo.everyFormatInfo[TotalVarDataIndex].varNum = uint32(lastVarNum)
	dataCamp.Reset()

	TotalVarDataIndex = TotalVarDataIndex + 1
	lastVarNum = len(RptVarList)
	lastAddr = div + lastAddr

	// 4.将变量信息写入结构体
	for i := 0; i < len(RptVarList); i++ {
		FinalRptFmtInfo.varInfo[i] = RptVarList[i]
	}
	// 5.转换为bin文件
	buffer := toBinaryDataRpt(FinalRptFmtInfo)
	binary.Write(buffer, binary.LittleEndian, totalbuffer.Bytes())

	if buffer.Len() < (fmtLen - 8) {
		for i := buffer.Len(); i < (fmtLen - 8); i++ {
			binary.Write(buffer, binary.LittleEndian, byte(fillchar))
		}
	}
	// 6.写结尾5a a5 a5 5a 00 00 00 00

	dataCamp1 := bytes.NewBufferString("")

	for i := 0; i < len(FMT_FILL_TAIL); i++ {
		dataCamp1.WriteByte(FMT_FILL_TAIL[i])
	}
	binary.Write(buffer, binary.LittleEndian, dataCamp1.Bytes())
	return buffer
}

// 转二进制
func toBinaryDataRpt(tempInfo printInfoRpt) *bytes.Buffer {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, tempInfo.printerName)
	binary.Write(buf, binary.LittleEndian, tempInfo.formatNum)
	binary.Write(buf, binary.LittleEndian, tempInfo.everyFormatInfo)
	binary.Write(buf, binary.LittleEndian, tempInfo.varInfo)
	return buf
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

func ParserDefFmtToFile(fmtDataList [12]string, printerModel string, fmtLen int) bool {
	var buffer *bytes.Buffer
	if printerModel == "EPM205" {
		buffer = ParserDefFmtToBuf(fmtDataList, printerModel, fmtLen)
	}
	// 7.创建bin文件
	if buffer.Len() > 0 {
		if creatFile("formatBin.bin", buffer) {
			fmt.Println("creat binary file success")
			return true
		} else {
			fmt.Println("creat binary file fail")
			return false
		}

	}
	return false

	// 8.写数据到串口
}

func ParserDefFmtToBuf(fmtDataList [12]string, printerModel string, fmtLen int) *bytes.Buffer {
	var FinalFormatInfo printInfoDef

	dataCamp1 := bytes.NewBufferString("")
	for i := 0; i < len(FMT_FILL_TAIL); i++ {
		dataCamp1.WriteByte(FMT_FILL_TAIL[i])
	}

	// var fillchar byte
	fillchar := 0xff
	var lastVarPos int //变量位置

	FinalFormatInfo.formatNum = 12 ///打印格式总数，根据打印格式文件数量决定

	totalbuffer := bytes.NewBufferString("") //打印命令集合
	// var everyAddr []int                                                 //每个打印格式偏移量
	var everyBufLen []int  //每个打印格式的命令集合
	TotalVarDataIndex := 0 //每个打印格式信息的索引
	// formatinfo.VarTable = formatinfo.ReadTableFromFile(currentPath + "\\varTable.json") //获取变量ID表
	dataCamp := bytes.NewBufferString("") //临时buf 存放命令数据
	lastVarNum := 0
	lastAddr := 0
	var clearList []VarStruct
	VarList = clearList // 用于清空数据

	//for循环解析文件
	for i := 0; i < 12; i++ {
		utf8Buff := fmtDataList[i]
		lastVarPos = 0

		buff, _ := Utf8ToGb2312(utf8Buff)
		formatbuf := ParseEplLines(buff, dataCamp, lastVarPos)
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
			lastAddr = headBufLenDef
			lastVarNum = len(VarList)
		}
		// fmt.Printf("%x", templen)
		FinalFormatInfo.everyFormatInfo[TotalVarDataIndex].addr = uint32(lastAddr)
		FinalFormatInfo.everyFormatInfo[TotalVarDataIndex].formatLen = uint32(everyBufLen[TotalVarDataIndex])
		FinalFormatInfo.everyFormatInfo[TotalVarDataIndex].varNum = uint32(lastVarNum)
		dataCamp.Reset()
		// fmt.Println(dataCamp.Len())
		TotalVarDataIndex = TotalVarDataIndex + 1
		lastVarNum = len(VarList)
		lastAddr = div + lastAddr

	}
	if len(VarList) > 150 {
		return bytes.NewBufferString("")

	}

	//复制变量
	copy(FinalFormatInfo.varInfo[:], VarList)

	buffer := binaryDataDef(FinalFormatInfo)
	binary.Write(buffer, binary.LittleEndian, totalbuffer.Bytes())
	// fmt.Println(buffer.Len())
	if buffer.Len() < (fmtLen - 8) {
		for i := buffer.Len(); i < (8192 - 8); i++ {
			binary.Write(buffer, binary.LittleEndian, byte(fillchar))
		}
	} else {
		return bytes.NewBufferString("")
	}
	binary.Write(buffer, binary.LittleEndian, dataCamp1.Bytes())

	return buffer

}

func binaryDataDef(tempInfo printInfoDef) *bytes.Buffer {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, tempInfo.printerName)
	binary.Write(buf, binary.LittleEndian, tempInfo.formatNum)
	binary.Write(buf, binary.LittleEndian, tempInfo.everyFormatInfo)
	binary.Write(buf, binary.LittleEndian, tempInfo.varInfo)
	return buf
}
