package svc

import (
	l "tmaxsrv/log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type DbModeSetting struct {
	dbName string
}

type ModeSetting struct {
	//ScaleMode 0 =DC500 1=check Weigher 2=take in  3=take out
	Id            uint `gorm:"primaryKey;autoincrement;not null"`
	ScaleMode     uint
	DateFormat    string
	DateSeparator string
	StableTime    string
	ZeroRange     string
	RecMode       string
	ScaleSn       string
	SaveMode      string
	WgtMode       uint `gorm:"default:0"` //0 单台模式  1 合并称重模式

}

func NewDbModeSetting(dbName string) (*DbModeSetting, error) {
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
	if err = db.AutoMigrate(&ModeSetting{}); err != nil {
		l.Log.Debug("failed to migrate database of scale connection")
	}

	return &DbModeSetting{dbName: dbName}, nil
}

func (d *DbModeSetting) GetModeSetting(scaleMode uint) ([]ModeSetting, error) {
	defer func() { recover() }()
	var err error
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
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
	if err = db.AutoMigrate(&ModeSetting{}); err != nil {
		l.Log.Debug("failed to migrate database of scale connection")
	}

	// 读取内容
	var setting []ModeSetting
	db.Where("scale_mode=?", scaleMode).Find(&setting)
	return setting, nil
}

func (d *DbModeSetting) UpdateModeSetting(setting ModeSetting) error {
	defer func() { recover() }()
	var err error
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
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
	// 明确指定要更新的字段
	updates := map[string]interface{}{
		"date_format":    setting.DateFormat,
		"date_separator": setting.DateSeparator,
		"stable_time":    setting.StableTime,
		"zero_range":     setting.ZeroRange,
		"rec_mode":       setting.RecMode,
		"scale_sn":       setting.ScaleSn,
		"save_mode":      setting.SaveMode,
		"wgt_mode":       setting.WgtMode,
	}
	rowAffected := db.Model(&ModeSetting{}).Where("scale_mode=?", setting.ScaleMode).Updates(updates).RowsAffected
	if rowAffected == 0 {
		// 如果不存在记录，则插入新记录以保存配置
		if err := db.Create(&setting).Error; err != nil {
			return err
		}
	}
	return nil
}
