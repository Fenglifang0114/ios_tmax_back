package eeprom

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"tmaxsrv/comm"

	"github.com/xuri/excelize/v2"
)

type EepromStruct struct {
	EepromStruct []EepromField
}
type EepromField struct {
	FieldName   string
	Size        int
	Addr        int
	Type        string
	Permission  int
	Values      string
	Comments    string
	Description string
}

const (
	TMAX_EEPROM_EXCEL = "T-MAXeeprom.xlsx"
)

func getExcelPath() string {

	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))
	currentPath := path[:index]
	currentPath = filepath.Join(currentPath, comm.SRV_DATA_PATH)
	currentPath = currentPath + "\\" + TMAX_EEPROM_EXCEL // 获取EEprom值
	fmt.Println(currentPath)
	return currentPath

}

// 读取excel文件
func getExcelData() (EepromStruct, bool) {

	var fieldData EepromStruct
	var filename = getExcelPath()
	f, err := excelize.OpenFile(filename)
	if err != nil {
		fmt.Println("read excel file error", err.Error())
		return fieldData, false
	}
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		fmt.Println("read excel file error", err.Error())
		return fieldData, false
	}
	for i, row := range rows {
		if i > 1 && len(row) >= 14 && strings.Contains(row[4], "M") {
			var tempEepromField EepromField
			tempEepromField.FieldName = row[4]
			size, err := strconv.Atoi(row[5])
			if err != nil {
				size = 0
			}
			tempEepromField.Size = size
			tempEepromField.Comments = row[8]
			addr, err := strconv.Atoi(row[9])
			if err != nil {
				addr = 0
			}
			tempEepromField.Addr = addr
			tempEepromField.Type = row[10]
			permission, err := strconv.Atoi(row[11])
			if err != nil {
				permission = 0
			}
			tempEepromField.Permission = permission
			tempEepromField.Values = row[12]
			tempEepromField.Description = row[13]
			if tempEepromField.Size != 0 {
				fieldData.EepromStruct = append(fieldData.EepromStruct, tempEepromField)
			}
		}
	}
	return fieldData, true
}

func GetFuncAddrSize(fieldName string) (addr int, size int) {

	excelData, res := getExcelData()
	if !res {
		return 0, 0
	}

	for _, data := range excelData.EepromStruct {
		// 找到满足条件的数据
		if data.FieldName == fieldName {
			return data.Addr, data.Size
		}
	}

	return 0, 0

}
