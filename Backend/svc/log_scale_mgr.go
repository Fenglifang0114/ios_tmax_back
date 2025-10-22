package svc

import (
	"fmt"
	"strings"
	"time"
)

type AddComInfo struct {
	IsDc500    bool
	SerialPort string
	Baud       int
	DataBits   int
	StopBits   int
	Parity     string
}

type UpdateComInfo struct {
	SerialPort string
	Baud       int
}

func SaveAddScaleLog(payload ReqAddScale) {

	jsonData := ""

	if payload.MediaConf.Type == MEDIA_COM {
		comInfo := ComInfo{}
		json.Unmarshal([]byte(payload.MediaConf.MediaInfoJson), &comInfo)

		addComInfo := AddComInfo{
			IsDc500:    payload.ScaleModel == "DC500",
			SerialPort: comInfo.DevPath,
			Baud:       comInfo.Baud,
			DataBits:   8,
			StopBits:   1,
			Parity:     "None",
		}
		jsonData, _ = json.MarshalToString(addComInfo)
	}

	if payload.MediaConf.Type == MEDIA_NET {
		netInfo := NetInfo{}
		json.Unmarshal([]byte(payload.MediaConf.MediaInfoJson), &netInfo)

		jsonData, _ = json.MarshalToString(netInfo)

	}

	LogSysOperation(MenuScalesManage, SubAdd, OpAddStr, jsonData, "ok", "")

}

func SaveModifyScaleNameLog(newName, oldName string) {
	jsonData := fmt.Sprintf(`{"NewName": "%s", "OldName": "%s"}`, newName, oldName)
	LogSysOperation(MenuScalesManage, SubUpdate, OpUpdateStr, jsonData, "ok", "")
}

func SaveModifyScaleLog(newMedia MediaConf, oldMedia MediaConf, scaleName string) {

	newUpdate := UpdateComInfo{}
	oldUpdate := UpdateComInfo{}
	if newMedia.Type == MEDIA_COM {
		newComInfo := ComInfo{}
		json.Unmarshal([]byte(newMedia.MediaInfoJson), &newComInfo)
		oldComInfo := ComInfo{}
		json.Unmarshal([]byte(oldMedia.MediaInfoJson), &oldComInfo)

		newUpdate = UpdateComInfo{
			SerialPort: newComInfo.DevPath,
			Baud:       newComInfo.Baud,
		}
		oldUpdate = UpdateComInfo{
			SerialPort: oldComInfo.DevPath,
			Baud:       oldComInfo.Baud,
		}
	}

	jsonData, _ := json.MarshalToString(map[string]interface{}{
		"ScaleName": scaleName,
		"NewInfo":   newUpdate,
		"OldInfo":   oldUpdate,
	})

	LogSysOperation(MenuScalesManage, SubUpdate, OpUpdateStr, jsonData, "ok", "")

}

func SaveDelScaleInfoLog(delScaleInfo MediaConf, delScaleName string) {
	jsonData := ""
	if delScaleInfo.Type == MEDIA_COM {
		comInfo := ComInfo{}
		json.Unmarshal([]byte(delScaleInfo.MediaInfoJson), &comInfo)
		jsonData, _ = json.MarshalToString(UpdateComInfo{
			SerialPort: comInfo.DevPath,
			Baud:       comInfo.Baud,
		})
	}
	if delScaleInfo.Type == MEDIA_NET {
		netInfo := NetInfo{}
		json.Unmarshal([]byte(delScaleInfo.MediaInfoJson), &netInfo)
		jsonData, _ = json.MarshalToString(netInfo)
	}
	jsonStr, _ := json.MarshalToString(map[string]interface{}{
		"ScaleName": delScaleName,
		"ScaleInfo": jsonData,
	})
	LogSysOperation(MenuScalesManage, SubDel, OpDeleteStr, jsonStr, "ok", "")
}

// 设置秤上的时间
func SaveSetScaleTimeLog(scaleName string, num uint64) {
	//将时间戳1760603041转为年月日时分秒
	t := time.Unix(int64(num), 0)
	year, month, day := t.Date()
	hour, min, sec := t.Clock()
	jsonData := fmt.Sprintf(`{"ScaleName": "%s", "Time": "%d-%02d-%02d %02d:%02d:%02d"}`, scaleName, year, month, day, hour, min, sec)
	LogSysOperation(MenuDeviceTime, SubDeviceTimeSet, OpSetStr, jsonData, "ok", "")
}

// 设置秤上的蓝牙名字
func SaveUpdateBtNameLog(scaleName string, name string) {
	jsonData := fmt.Sprintf(`{"ScaleName": "%s", "BtName": "%s"}`, scaleName, name)
	LogSysOperation(MenuDeviceBt, SubBtSetName, OpSetStr, jsonData, "", "")
}

func SaveUpdateBtPowerLog(scaleName string, data string) {
	hasRFPower := strings.Contains(data, "RFPOWER")
	power := "Unknown"
	if hasRFPower {
		switch {
		case strings.Contains(data, "14"):
			power = "Strong"
		case strings.Contains(data, "11"):
			power = "Medium"
		case strings.Contains(data, "5"):
			power = "Weak"
		}
	}
	jsonData := fmt.Sprintf(`{"ScaleName": "%s", "BtPower": "%s"}`, scaleName, power)
	LogSysOperation(MenuDeviceBt, SubBtSetPower, OpSetStr, jsonData, "", "")
}

// 设置秤上的固件版本
func SaveUpdateFirmwareLog(scaleName string, firmware string, mode string) {
	jsonData := fmt.Sprintf(`{"ScaleName": "%s", "FirmwareFile": "%s" , "Mode": "%s"}`, scaleName, firmware, mode)
	LogSysOperation(MenuUpdateFirmware, SubUpdateFirmware, OpUpdateStr, jsonData, "", "")
}

// 保存下发标签打印格式到日志
func SaveDownLabelFmtToScaleLog(scaleName string, reqData string) {
	jsonData := fmt.Sprintf(`{"ScaleName": "%s", "FilePath": "%s"}`, scaleName, reqData)
	LogSysOperation(MenuDownLabelFmt, SubDownLabelFmt, OpIssueStr, jsonData, "ok", "")

}

// 保存下发打印格式到日志
func SaveDownRptFmtToScaleLog(scaleName string, reqData string) {
	jsonData := fmt.Sprintf(`{"ScaleName": "%s", "FilePath": "%s"}`, scaleName, reqData)
	LogSysOperation(MenuDownReceiptFmt, SubDownReceiptFmt, OpIssueStr, jsonData, "ok", "")

}

// 保存新增称重记录到日志
func SaveAddScaleWgtLog(payload ReqAddWgtRec) {
	LogScaleWgtOperation(0, payload.HeadRec, "")
}

// 保存清空称重记录的日志

func SaveClearScaleWgtLog(mode int) {
	module := GetScaleWgtMode(mode)
	LogSysOperation(module, module, OpClearStr, "", "ok", "")
}
