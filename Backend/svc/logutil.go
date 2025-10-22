package svc

import (
	"context"

	"time"
)

// // 操作类型ID常量定义（与数据库OperationType表对应）
// const (
// 	OpTypeLogin  = 1
// 	OpTypeLogout = 2
// 	OpTypeCreate = 3
// 	OpTypeDelete = 4
// 	OpTypeUpdate = 5
// 	OpTypeQuery  = 6
// 	OpTypeIssue  = 7 // 下发
// 	OpTypeImport = 8
// 	OpTypeExport = 9
// 	OpTypeOther  = 10
// )

// // 菜单ID常量定义（与数据库OperationMenu表对应）
// const (
// 	MenuPrintFormat = 1 // 假设打印格式菜单ID为1
// 	MenuScaleManage = 2
// 	MenuPLUList     = 3
// 	// ... 其他菜单ID
// )

// LogOperation 统一操作日志记录接口
func LogOperation(ctx context.Context, opTypeID int, menuID int, content string, status int, result string, deviceID int, deviceSn string) {
	// 1. 获取当前用户信息
	userID, username, roleID := GetCurrentUser()
	if userID == 0 && ctx != nil {
		// 尝试从Context获取（适用于HTTP请求场景）
		if userInfo, ok := ctx.Value(UserContextKey{}).(UserInfo); ok {
			userID = userInfo.UserID
			username = userInfo.Username
			roleID = userInfo.RoleId
		}
	}

	// 2. 构造操作日志对象
	log := &OperationLog{
		OperatorTypeID:  opTypeID,
		OperatorMenuID:  menuID,
		OperatorContent: content,
		Status:          status,
		Result:          result,
		DeviceId:        deviceID,
		DeviceSn:        deviceSn,
		OperatorId:      userID,
		OperatorName:    username,
		RoleId:          roleID,
		OperationTime:   time.Now(),
	}

	println("LogOperation: ", log)

	// 3. 异步写入数据库（不阻塞主流程）
	// go func() {
	// 	// 获取数据库实例（假设全局DBOperationLog实例已初始化）
	// 	if err := dbOperationLog.CreateOperationLog(log); err != nil {
	// 		l.Log.Error(fmt.Sprintf("Failed to write operation log: %v, log content: %+v", err, log))
	// 	} else {
	// 		l.Log.Debug(fmt.Sprintf("Operation log written successfully: %+v", log))
	// 	}
	// }()
}

// // 5. 记录操作日志（异步执行） 在请求的函数中实现
// LogOperation(req.Ctx, OpTypeIssue, MenuPrintFormat,
// 	fmt.Sprintf("下发打印格式: ID=%s, 格式名称=%s, 内容=%s",
// 		req.FormatID, req.FormatName, req.FormatContent),
// 	status, result, c.DeviceId, c.DeviceSn)
