package svc

import (
	"errors"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DbProductRec struct {
	dbName string
}

type ProductRec struct {
	RecId       uint `gorm:"primaryKey;autoincrement;not null"`
	Id          string
	Product     string
	WithPretare bool
	Pretare     string
	Remarks     string
	CreatedAt   time.Time
}

func NewDbProductRec(dbName string) (*DbProductRec, error) {
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
	if err = db.AutoMigrate(&ProductRec{}); err != nil {
		panic("failed to migrate database of scale connection")
	}
	return &DbProductRec{dbName: dbName}, nil
}

func (d *DbProductRec) GetProductRecsList() ([]ProductRec, error) {
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
	if err = db.AutoMigrate(&ProductRec{}); err != nil {
		panic("failed to migrate database of product")
	}

	// 读取内容
	var recs []ProductRec
	db.Find(&recs)
	return recs, nil
}

func (d *DbProductRec) GetProductRecs(start int, quantity int) ([]ProductRec, error) {
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
	if err = db.AutoMigrate(&ProductRec{}); err != nil {
		panic("failed to migrate database of scale connection")
	}

	// 读取内容
	var recs []ProductRec
	db.Limit(quantity).Offset(start).Find(&recs)
	return recs, nil
}

func (d *DbProductRec) InsertProductRec(rec ProductRec) error {
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

func (d *DbProductRec) UpdateProductRec(rec ProductRec) error {
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
		return errors.New("@UpdateProductRec failed, mybe record not existing")
	}
	return nil
}

func (d *DbProductRec) DeleteProductRec(id uint) error {
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
	var rec ProductRec
	rec.RecId = id
	db.Where("rec_id=?", id).Delete((&rec))
	return nil
}
