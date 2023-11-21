package svc

import (
	"errors"

	"gorm.io/driver/sqlite"
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
}

func NewDbModeSetting(dbName string) (*DbModeSetting, error) {
	var err error
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	sqlDB, err := db.DB()
	if err != nil {
		panic("failed to connect database")
	}
	if sqlDB != nil {
		defer sqlDB.Close()
	}
	// Migrate the schema
	if err = db.AutoMigrate(&ModeSetting{}); err != nil {
		panic("failed to migrate database of scale connection")
	}
	return &DbModeSetting{dbName: dbName}, nil
}

func (d *DbModeSetting) GetModeSetting(scaleMode uint) ([]ModeSetting, error) {
	var err error
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	sqlDB, err := db.DB()
	if err != nil {
		panic("failed to connect database")
	}
	if sqlDB != nil {
		defer sqlDB.Close()
	}
	// Migrate the schema
	if err = db.AutoMigrate(&ModeSetting{}); err != nil {
		panic("failed to migrate database of scale connection")
	}

	// 读取内容
	var setting []ModeSetting
	db.Where("scale_mode=?", scaleMode).Find(&setting)
	return setting, nil
}

func (d *DbModeSetting) UpdateModeSetting(setting ModeSetting) error {
	var err error
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	sqlDB, err := db.DB()
	if err != nil {
		panic("failed to connect database")
	}
	if sqlDB != nil {
		defer sqlDB.Close()
	}
	rowAffected := db.Model(&setting).Where("scale_mode=?", setting.ScaleMode).Updates(&setting).RowsAffected
	if rowAffected == 0 {
		return errors.New("@UpdateScaleRec failed, mybe record not existing")
	}
	return nil
}
