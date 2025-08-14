package svc

import (
	"errors"
	"time"
	l "tmaxsrv/log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DbWifiRec struct {
	dbName string
}

type WifiRec struct {
	Id        uint `gorm:"primaryKey;autoincrement;not null"`
	Ssid      string
	Pwd       string
	CreatedAt time.Time
}

func NewDbWifiRec(dbName string) (*DbWifiRec, error) {
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
	if err = db.AutoMigrate(&WifiRec{}); err != nil {
		l.Log.Debug("failed to migrate database of scale connection")
	}
	return &DbWifiRec{dbName: dbName}, nil
}

func (d *DbWifiRec) GetWifiRecsList() ([]WifiRec, error) {
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
	if err = db.AutoMigrate(&WifiRec{}); err != nil {
		l.Log.Debug("failed to migrate database of scale connection")
	}

	// 读取内容
	var recs []WifiRec
	db.Find(&recs)
	return recs, nil
}

func (d *DbWifiRec) GetWifiRecs(start int, quantity int) ([]WifiRec, error) {
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
	if err = db.AutoMigrate(&WifiRec{}); err != nil {
		l.Log.Debug("failed to migrate database of scale connection")
	}

	// 读取内容
	var recs []WifiRec
	db.Limit(quantity).Offset(start).Find(&recs)
	return recs, nil
}

func (d *DbWifiRec) InsertWifiRec(rec WifiRec) error {
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

	db.Where("ssid=?", rec.Ssid).Delete((&rec)) //先删除一样的

	tx := db.Create(&rec)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (d *DbWifiRec) UpdateWifiRec(rec WifiRec) error {
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
	var oldRec WifiRec
	if err := db.Where("ssid = ?", rec.Ssid).First(&oldRec).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("@UpdateWifiRec failed, mybe record not existing")
		} else {
			return errors.New("@UpdateWifiRec failed, mybe record not existing")
		}
	}
	db.Save(&rec)
	return nil
}

func (d *DbWifiRec) DeleteWifiRec(id uint) error {
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

	var rec WifiRec
	rec.Id = id
	db.Where("rec_id=?", id).Delete((&rec))
	return nil
}
