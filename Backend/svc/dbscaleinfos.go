package svc

import (
	"errors"
	"time"
	l "tmaxsrv/log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DbScaleInfos struct {
	dbName string
}

type ScaleInfos struct {
	RecId      uint `gorm:"primaryKey;autoincrement;not null"`
	ScaleModel string
	ScaleSn    string
	ScaleName  string
	CreatedAt  time.Time
}

func NewDbScaleInfos(dbName string) (*DbScaleInfos, error) {
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
	if err = db.AutoMigrate(&ScaleInfos{}); err != nil {
		l.Log.Debug("failed to migrate database of scale infos")
	}
	return &DbScaleInfos{dbName: dbName}, nil
}

func (d *DbScaleInfos) GetScaleInfo(model string, sn string) ([]ScaleInfos, error) {
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
	if err = db.AutoMigrate(&ScaleInfos{}); err != nil {
		l.Log.Debug("failed to migrate database of scale infos")
	}

	// 读取内容
	var infos []ScaleInfos
	db.Where("scale_model=?", model).Where("scale_sn=?", sn).Find(&infos)
	return infos, nil
}

func (d *DbScaleInfos) InsertScaleInfos(rec ScaleInfos) error {
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

	tx := db.Create(&rec)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (d *DbScaleInfos) UpdateScaleInfos(rec ScaleInfos) error {
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
	rowAffected := db.Model(&rec).Where("rec_id=?", rec.RecId).Updates(&rec).RowsAffected
	if rowAffected == 0 {
		return errors.New("@UpdateScaleInfos failed, mybe record not existing")
	}
	return nil
}

func (d *DbScaleInfos) DeleteScaleInfos(id uint) error {
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

	var info ScaleInfos
	info.RecId = id
	db.Where("rec_id=?", id).Delete((&info))
	return nil
}

func (d *DbScaleInfos) DeleteAllScaleInfos() error {
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

	err = db.Migrator().DropTable(&ScaleInfos{})
	if err != nil {
		l.Log.Debug("failed to drop database")
	}

	return nil
}
