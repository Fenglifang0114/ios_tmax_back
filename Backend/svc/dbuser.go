package svc

import (
	"errors"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DbUserRec struct {
	dbName string
}

type UserRec struct {
	RecId     uint `gorm:"primaryKey;autoincrement;not null"`
	Id        string
	Name      string
	IsFemale  bool
	Phone     string
	Remarks   string
	CreatedAt time.Time
}

func NewDbUserRec(dbName string) (*DbUserRec, error) {
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
	return &DbUserRec{dbName: dbName}, nil
}

func (d *DbUserRec) GetUserRecsList() ([]UserRec, error) {
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
	if err = db.AutoMigrate(&UserRec{}); err != nil {
		panic("failed to migrate database of scale connection")
	}

	// 读取内容
	var recs []UserRec
	db.Find(&recs)
	return recs, nil
}

func (d *DbUserRec) GetUserRecs(start int, quantity int) ([]UserRec, error) {
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
	if err = db.AutoMigrate(&UserRec{}); err != nil {
		panic("failed to migrate database of scale connection")
	}

	// 读取内容
	var recs []UserRec
	db.Limit(quantity).Offset(start).Find(&recs)
	return recs, nil
}

func (d *DbUserRec) InsertUserRec(rec UserRec) error {
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

func (d *DbUserRec) UpdateUserRec(rec UserRec) error {
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
	rowAffected := db.Model(&rec).Where("rec_id=?", rec.RecId).Updates(&rec).RowsAffected
	if rowAffected == 0 {
		return errors.New("@UpdateUserRec failed, mybe record not existing")
	}
	return nil
}

func (d *DbUserRec) DeleteUserRec(id uint) error {
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

	if err != nil {
		panic("failed to connect database")
	}
	var rec UserRec
	rec.RecId = id
	db.Where("rec_id=?", id).Delete((&rec))
	return nil
}
