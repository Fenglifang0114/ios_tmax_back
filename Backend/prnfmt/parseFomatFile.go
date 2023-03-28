package prnfmt

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/xuri/excelize/v2"
)

func GetFormatLines(fileNames string) ([]string, bool) {
	var formatArray []string
	res := false
	lineIdx := 0

	f, err := os.Open(fileNames)
	if err != nil {
		panic("open file error")

	}
	defer f.Close() //打开文件出错处理
	if nil == err {
		buff := bufio.NewReader(f) //读入缓存
		for {

			line, err := buff.ReadString('\n') //以'\n'为结束符读入一行
			if err != nil || io.EOF == err {
				break
			}
			if lineIdx == 0 && strings.Contains(line, "\ufeff") {
				line = strings.Replace(line, "\ufeff", "", 1)

			}
			formatArray = append(formatArray, line)
		}

	}
	if len(formatArray) > 1 {
		res = true
	}
	return formatArray, res
}

func ReadTableFromFile(tableFilePath string) ScaleVarOrder {
	jsonFile, err := os.Open(tableFilePath)
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	jsonfile, _ := ioutil.ReadAll(jsonFile)
	// fmt.Print(byteValue)
	var tempTable ScaleVarOrder
	err = json.Unmarshal(jsonfile, &tempTable)
	if err != nil {
		log.Panic(err)
	}
	return tempTable
}

//读取excel文件，返回行列数组

func getBarCodeTypeId(filename string, tempLan string, tempType string) string {

	ret := ""
	f, err := excelize.OpenFile(filename)
	if err != nil {
		fmt.Println("读取excel文件出错", err.Error())
		return ret
	}
	// sheets := f.GetSheetMap()
	// fmt.Println(sheets)
	// sheet1 := sheets[1]
	// fmt.Println("第一个工作表", sheet1)
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		fmt.Println("读取excel文件出错", err.Error())
		return ret
	}
	cols := []string{}
	isOk := false

	for i, row := range rows {
		if i == 0 { //取得第一行的所有数据---execel表头
			cols = append(cols, row...)
			isOk = true
			// fmt.Println("列信息", cols)

		} else {
			if isOk {
				isOk = false
				break
			}
		}
	}
	colId := findRowColId(cols, tempLan)

	row1 := []string{}
	rowDes := []string{}
	isDesOk := false
	if colId > 0 {
		cols1, err := f.GetCols("Sheet1")
		if err != nil {
			fmt.Println("读取excel文件出错", err.Error())
			return ret
		}

		for i, col := range cols1 {
			if i == 1 { //取得第一行的所有数据---execel表头
				row1 = append(row1, col...)
				isOk = true
				// fmt.Println("行信息", row1)

			} else if i == colId {
				rowDes = append(rowDes, col...)
				isDesOk = true
				// fmt.Println("行信息", rowDes)

			} else {
				if isOk && isDesOk {
					break
				}
			}
		}
	}

	rowId := findRowColId(row1, tempType)
	if rowId > 0 {
		ret = rowDes[rowId]
	}
	return ret
}

func findRowColId(tempArr []string, tempVar string) int {
	id := 0
	for i := 0; i < len(tempArr); i++ {
		if tempArr[i] == tempVar {
			id = i
			break
		}
	}
	return id
}
