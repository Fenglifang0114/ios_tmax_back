package svc

import (
	"errors"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DbScaleRec struct {
	dbName string
}

type ScaleRec struct {
	RecId       uint `gorm:"primaryKey;autoincrement;not null"`
	ScaleModel  string
	ScaleSn     string
	Product     string
	Weight      string
	Price       string
	PluNo       string
	PluRemarks  string
	WeightUnit  string
	Pretare     string
	UserNo      string
	UserName    string
	UserRemarks string
	ScaleMode   string //0 =DC500 1=check Weigher 2=take in  3=take out

	CreatedAt time.Time
}

func NewDbScaleRec(dbName string) (*DbScaleRec, error) {
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
	if err = db.AutoMigrate(&ScaleRec{}); err != nil {
		panic("failed to migrate database of scale connection")
	}
	return &DbScaleRec{dbName: dbName}, nil
}

func (d *DbScaleRec) GetScaleRecsList(model string, sn string) ([]ScaleRec, error) {
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
	if err = db.AutoMigrate(&ScaleRec{}); err != nil {
		panic("failed to migrate database of scale connection")
	}

	// 读取内容
	var recs []ScaleRec
	db.Where("scale_model=?", model).Where("scale_sn=?", sn).Find(&recs)
	return recs, nil
}

func (d *DbScaleRec) GetScaleRecs(model string, sn string, start int, quantity int) ([]ScaleRec, error) {
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
	if err = db.AutoMigrate(&ScaleRec{}); err != nil {
		panic("failed to migrate database of scale connection")
	}

	// 读取内容
	var recs []ScaleRec
	db.Limit(quantity).Offset(start).Where("scale_model=?", model).Where("scale_sn=?", sn).Find(&recs)
	return recs, nil
}

func (d *DbScaleRec) InsertScaleRec(rec ScaleRec) error {
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

func (d *DbScaleRec) UpdateScaleRec(rec ScaleRec) error {
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
		return errors.New("@UpdateScaleRec failed, mybe record not existing")
	}
	return nil
}

func (d *DbScaleRec) DeleteScaleRec(id uint) error {
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
	var rec ScaleRec
	rec.RecId = id
	db.Where("rec_id=?", id).Delete((&rec))
	return nil
}

func (d *DbScaleRec) DeleteAllScaleRec() error {
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

	err = db.Migrator().DropTable(&ScaleRec{})
	if err != nil {
		panic("failed to drop database")
	}

	return nil
}
