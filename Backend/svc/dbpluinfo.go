package svc

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DbPluRec struct {
	dbName string
}

type PluRec struct {
	RecId     uint `gorm:"primaryKey;autoincrement;not null"`
	Md5       string
	FileName  string
	CreatedAt time.Time
}

func NewDbPluRec(dbName string) (*DbPluRec, error) {
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
	if err = db.AutoMigrate(&UserRec{}); err != nil {
		panic("failed to migrate database of scale connection")
	}
	return &DbPluRec{dbName: dbName}, nil
}

func (d *DbPluRec) GetPluRecsList() ([]PluRec, error) {
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
	if err = db.AutoMigrate(&PluRec{}); err != nil {
		panic("failed to migrate database of scale connection")
	}

	// 读取内容
	var recs []PluRec
	db.Find(&recs)
	return recs, nil
}

func (d *DbPluRec) GetPluRecs(md5Str string) ([]PluRec, error) {
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
	if err = db.AutoMigrate(&PluRec{}); err != nil {
		panic("failed to migrate database of scale connection")
	}

	// 读取内容
	var recs []PluRec
	db.Where("md5=?", md5Str).Order("created_at desc").First(&recs)
	return recs, nil
}

func (d *DbPluRec) InsertPluRec(rec PluRec) error {
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

	tx := db.Create(&rec)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}
