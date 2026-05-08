package svc

import (
	"time"
	l "tmaxsrv/log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

//系统用户表

type DbOperationLog struct {
	dbName string
}

type OperationLog struct {
	RecId           int `gorm:"primaryKey;not null;autoincrement;"`
	OperatorTypeID  int // 关联operation_types.ID
	OperatorMenuID  int // 关联operation_menus.ID
	OperatorContent string
	Status          int
	Result          string

	DeviceId    int
	DeviceModel string
	DeviceSn    string

	Remark string

	OperationTime time.Time `gorm:"autoCreateTime"`
	OperatorId    int       `gorm:"not null;"`
	OperatorName  string    `gorm:"not null;"`
	RoleId        int       `gorm:"not null;"`

	// 关联关系（可选，用于GORM查询）
	OperatorType *OperationType `gorm:"foreignKey:OperatorTypeID"`
	OperatorMenu *OperationMenu `gorm:"foreignKey:OperatorMenuID"`
}

type OperationType struct {
	ID   int    `gorm:"primaryKey"`
	Name string `gorm:"not null;unique"` // 如"登录"、"增加"
	Desc string // 描述（可选）
}

// 操作类型 1：登录 2：退出 3：增加 4：删除 5：修改 6：查询 7：下发 8：导入 9：导出 10：其他

type OperationMenu struct {
	ID       int    `gorm:"primaryKey"`
	Name     string `gorm:"not null;unique"` // 如"秤管理"、"PLU列表"
	Path     string // 菜单路径（可选，如"/scale/manage"）
	ParentID int    // 父菜单ID（如果有层级关系）
}

func NewDbOperationLog(dbName string) (*DbOperationLog, error) {
	defer func() { recover() }()
	var err error
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	if sqlDB != nil {
		defer sqlDB.Close()
	}
	// Migrate the schema
	if err = db.AutoMigrate(
		&OperationType{},
		&OperationMenu{},
		&OperationLog{},
	); err != nil {
		l.Log.Debug("failed to migrate database of scale connection")
	}

	return &DbOperationLog{dbName: dbName}, nil
}

// 创建操作日志
func (d *DbOperationLog) CreateOperationLog(log *OperationLog) error {
	defer func() { recover() }()
	var err error
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	if sqlDB != nil {
		defer sqlDB.Close()
	}

	if err = db.Create(log).Error; err != nil {
		return err
	}
	return nil
}

// 删除操作日志
func (d *DbOperationLog) DeleteOperationLog(logID []int) error {
	defer func() { recover() }()
	var err error
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	if sqlDB != nil {
		defer sqlDB.Close()
	}
	for _, id := range logID {
		if err = db.Delete(&SysUser{}, "user_id = ?", id).Error; err != nil {
			return err
		}
	}

	return nil
}

// 查询操作日志列表
func (d *DbOperationLog) ListOperationLogs() ([]OperationLog, error) {
	defer func() { recover() }()
	var err error
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		return []OperationLog{}, err

	}
	sqlDB, err := db.DB()
	if err != nil {
		return []OperationLog{}, err
	}
	if sqlDB != nil {
		defer sqlDB.Close()
	}
	var logs []OperationLog
	result := db.Find(&logs)
	if result.Error != nil {
		return []OperationLog{}, err
	}
	if result.RowsAffected == 0 {
		return []OperationLog{}, err
	}
	return logs, result.Error
}
