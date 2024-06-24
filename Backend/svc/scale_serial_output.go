package svc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var idMap = map[string]byte{
	"Date":        1,
	"Time":        2,
	"NO":          7,
	"Gross":       8,
	"Tare":        9,
	"Net":         10,
	"PCS":         11,
	"UnitWeight":  12,
	"Percent":     13,
	"TotalWeight": 14,
	"WeightUnit":  15,
	"TotalCount":  16,
	"isstable":    31,
	"istare":      32,
	// 添加更多的varname和对应的ID，如 "varname": ID
}

var alignmentMap = map[string]byte{
	"left":   0,
	"right":  1,
	"center": 2,
	// 添加更多的varname和对应的ID，如 "varname": ID
}

var typeMap = map[string]byte{
	"free":    0,
	"bool":    1,
	"integer": 2,
	"float":   3,
	"string":  4,
	// 添加更多的varname和对应的ID，如 "varname": ID
}

var fillingMap = map[string]byte{
	"0": 0,
	" ": 1,
}

type Function struct {
	ID          int    `json:"id"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Alignment   string `json:"alignment,omitempty"`
	Filling     string `json:"filling,omitempty"`
	Decimal     byte   `json:"decimal,omitempty"`
	IsTrue      string `json:"istrue,omitempty"`
	IsFalse     string `json:"isfalse,omitempty"`
}

type Data struct {
	IsVar      bool   `json:"isvar"`
	Value      string `json:"value,omitempty"`
	VarName    string `json:"varname,omitempty"`
	FunctionID int    `json:"functionid,omitempty"`
	Length     int    `json:"length,omitempty"`
	IsHex      bool   `json:"ishex,omitempty"`
}

type JSONData struct {
	Functions []Function `json:"function"`
	Data      []Data     `json:"data"`
}

type FunctionMap struct {
	ID       int
	FuncData []byte
}

func ParserSerialOutputFile(fileContext string) (bytes.Buffer, bool) {
	println(fileContext)
	jsonData := fileContext
	var sendArray bytes.Buffer //整包数据
	sendArray.Reset()

	var jsonDataStruct JSONData
	err := json.Unmarshal([]byte(jsonData), &jsonDataStruct)
	if err != nil {
		fmt.Println("解析JSON失败:", err)
		return sendArray, false
	}
	// 构建FunctionMap
	functionMapList := procJsonDataFunction(jsonDataStruct)
	// 构建function 数组和 Function的起始位置与id的对应关系
	funcArray, funcPosArray := writeToByteArray(functionMapList)
	//构建数据
	outputDataArray, varPosArray := procJsonDataData(jsonDataStruct, funcPosArray)

	//汇总数据   数据总长度2个字节，位置总长度2个字节，方法首地址2个字节，数据首地址两个字节
	totalPosLenByte := make([]byte, 2)
	binary.LittleEndian.PutUint16(totalPosLenByte, uint16(len(varPosArray)))
	totalFuncLenByte := make([]byte, 2)
	binary.LittleEndian.PutUint16(totalFuncLenByte, uint16(len(funcArray)))
	totalDateByte := make([]byte, 2)
	binary.LittleEndian.PutUint16(totalDateByte, uint16(len(outputDataArray)))
	totalLenByte := make([]byte, 2)
	binary.LittleEndian.PutUint16(totalLenByte, uint16(8+len(varPosArray)+len(funcArray)+len(outputDataArray)))

	totalFuncPosByte := make([]byte, 2)
	binary.LittleEndian.PutUint16(totalFuncPosByte, uint16(len(varPosArray)+8))
	totalDataPosByte := make([]byte, 2)
	binary.LittleEndian.PutUint16(totalDataPosByte, uint16(len(varPosArray)+len(funcArray)+8))

	sendArray.Write(totalLenByte)
	sendArray.Write(totalPosLenByte)
	sendArray.Write(totalFuncPosByte)
	sendArray.Write(totalDataPosByte)
	sendArray.Write(varPosArray)
	sendArray.Write(funcArray)
	sendArray.Write(outputDataArray)
	// writeDataToBin(sendArray)
	return sendArray, true
}

func procJsonDataData(jsonDataStruct JSONData, funcPosArray map[int]int) ([]byte, []byte) {
	outputDataArray := []byte{}
	varPosArray := []byte{}
	for _, data := range jsonDataStruct.Data {
		if data.IsVar {
			id, exists := idMap[data.VarName]
			if !exists {
				fmt.Printf("未知的变量名: %s\n", data.VarName)
				continue
			}
			start, exists := funcPosArray[data.FunctionID]
			if !exists {
				fmt.Printf("未找到FunctionID %d 的起始位置\n", data.FunctionID)
				continue
			}
			length := data.Length
			positionData := make([]byte, 6)
			positionData[0] = id
			positionData[2] = byte(start >> 8)   // 取start的高位字节
			positionData[1] = byte(start & 0xFF) // 取start的低位字节
			positionData[3] = byte(length)
			// 填充0到第五和第六个字节
			positionData[4] = 0
			positionData[5] = 0
			lengthBytes := make([]byte, 2)
			binary.LittleEndian.PutUint16(lengthBytes, uint16(len(outputDataArray)))
			varPosArray = append(varPosArray, lengthBytes...)
			outputDataArray = append(outputDataArray, positionData...) // 添加6个字节的空间
		} else {

			if data.IsHex {
				var result []byte

				parts := strings.Fields(data.Value)
				for _, part := range parts {
					value, _ := strconv.ParseInt(part, 16, 0)
					result = append(result, byte(value))
				}
				outputDataArray = append(outputDataArray, result...)

			} else {
				outputDataArray = append(outputDataArray, []byte(data.Value)...) // 追加value值

			}

		}
	}
	return outputDataArray, varPosArray
}

func procJsonDataFunction(jsonDataStruct JSONData) []FunctionMap {
	functionMapList := []FunctionMap{}
	for _, function := range jsonDataStruct.Functions {
		funcData := []byte{}

		funcType, exists := typeMap[function.Type] // 查找typeMap中对应的字节
		if !exists {
			fmt.Println("未知的函数类型:", function.Type)
		}

		funcTypeByte := []byte{funcType} // 将funcType转换为字节数组0
		if function.Type == "free" {
			funcData = append(funcData, funcTypeByte...)
			functionMapList = append(functionMapList, FunctionMap{
				ID:       function.ID,
				FuncData: funcData,
			})
		} else if function.Type == "integer" {
			funcAlignment, exists := alignmentMap[function.Alignment] // 查找typeMap中对应的字节
			if !exists {
				fmt.Println("未知的对齐方式:", function.Type)
			}
			funcFilling, exists := fillingMap[function.Filling] // 查找typeMap中对应的字节
			if !exists {
				fmt.Println("未知的填充方式:", function.Type)
			}
			funcAlignByte := []byte{funcAlignment}
			funcFillingByte := []byte{funcFilling}
			funcData = append(funcData, funcTypeByte...)
			funcData = append(funcData, funcAlignByte...)
			funcData = append(funcData, funcFillingByte...)
			functionMapList = append(functionMapList, FunctionMap{
				ID:       function.ID,
				FuncData: funcData,
			})
		} else if function.Type == "float" {
			funcAlignment, exists := alignmentMap[function.Alignment] // 查找typeMap中对应的字节
			if !exists {
				fmt.Println("未知的对齐方式:", function.Type)
			}
			funcFilling, exists := fillingMap[function.Filling] // 查找typeMap中对应的字节
			if !exists {
				fmt.Println("未知的填充方式:", function.Type)
			}

			funcAlignByte := []byte{funcAlignment}
			funcFillingByte := []byte{funcFilling}
			funcDecimalByte := []byte{function.Decimal}

			funcData = append(funcData, funcTypeByte...)
			funcData = append(funcData, funcAlignByte...)
			funcData = append(funcData, funcFillingByte...)
			funcData = append(funcData, funcDecimalByte...)

			functionMapList = append(functionMapList, FunctionMap{
				ID:       function.ID,
				FuncData: funcData,
			})
		} else if function.Type == "bool" {
			funcIsTrueByte := []byte(function.IsTrue)
			funcIsFalseByte := []byte(function.IsFalse)
			totalLen := byte(len(funcIsTrueByte)) + byte(len(funcIsFalseByte))

			funcData = append(funcData, funcTypeByte...)
			funcData = append(funcData, totalLen)
			funcData = append(funcData, byte(len(funcIsTrueByte)))
			funcData = append(funcData, funcIsTrueByte...)
			funcData = append(funcData, funcIsFalseByte...)

			functionMapList = append(functionMapList, FunctionMap{
				ID:       function.ID,
				FuncData: funcData,
			})
		} else if function.Type == "string" {
			funcAlignment, exists := alignmentMap[function.Alignment] // 查找typeMap中对应的字节
			if !exists {
				fmt.Println("未知的对齐方式:", function.Type)
			}
			funcFilling, exists := fillingMap[function.Filling] // 查找typeMap中对应的字节
			if !exists {
				fmt.Println("未知的填充方式:", function.Type)
			}
			funcAlignByte := []byte{funcAlignment}
			funcFillingByte := []byte{funcFilling}
			funcData = append(funcData, funcTypeByte...)
			funcData = append(funcData, funcAlignByte...)
			funcData = append(funcData, funcFillingByte...)
			functionMapList = append(functionMapList, FunctionMap{
				ID:       function.ID,
				FuncData: funcData,
			})
		}
	}
	return functionMapList
}

func WriteDataToBin(sendArray bytes.Buffer) {
	// 将 sendArray 中的数据写入到 bin 文件中
	bufferData := sendArray.Bytes()
	// hexString := fmt.Sprintf("% X", bufferData)
	// fmt.Println(hexString)
	// 创建一个长度为 2K 的切片，用于存储写入 bin 文件的数据
	outputData := make([]byte, 2048)
	// 如果 bufferData 的长度大于 2K，则只取前 2K 部分；否则，将所有 bufferData 内容复制到 outputData 中并在尾部补 0
	dataLength := len(bufferData)
	if dataLength > 2048 {
		copy(outputData, bufferData[:2048])
	} else {
		copy(outputData, bufferData)
	}
	for i := dataLength; i < 2048; i++ {
		outputData[i] = 0xFF
	}
	// 将 outputData 写入 bin 文件中
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println(err)
	}
	// 构建 bin 文件的路径，即当前执行程序的路径下的 output.bin
	binPath := filepath.Join(filepath.Dir(exePath), "output.bin")
	// 将 outputData 写入 bin 文件中
	err = os.WriteFile(binPath, outputData, fs.FileMode(0644))
	if err != nil {
		fmt.Println(err)
	}
	if err != nil {
		fmt.Println(err)
	}
}
func writeToByteArray(functions []FunctionMap) ([]byte, map[int]int) {
	// 根据ID对FunctionMap进行排序
	sort.SliceStable(functions, func(i, j int) bool {
		return functions[i].ID < functions[j].ID
	})
	// 创建一个字节切片来存储所有的FuncData
	var result []byte
	// 创建一个映射来记录FuncData在字节切片中的起始位置
	startPositions := make(map[int]int)
	// 遍历FunctionMap，并将FuncData写入字节切片中
	currentPos := 0
	for _, fn := range functions {
		startPositions[fn.ID] = currentPos
		result = append(result, fn.FuncData...)
		currentPos += len(fn.FuncData)
	}
	return result, startPositions
}
