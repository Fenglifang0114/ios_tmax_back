package svc

import (
	"errors"

	"github.com/gitteamer/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DbScaleConn struct {
	dbName string
}

func NewDbScaleConn(dbName string) (*DbScaleConn, error) {
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
	if err = db.AutoMigrate(&ScaleConnMedia{}); err != nil {
		panic("failed to migrate database of scale connection")
	}
	return &DbScaleConn{dbName: dbName}, nil
}

func (d *DbScaleConn) GetScaleConnList() ([]*ScaleConnMedia, error) {
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
	if err = db.AutoMigrate(&ScaleConnMedia{}); err != nil {
		panic("failed to migrate database of scale connection")
	}

	// 读取内容
	var conns []*ScaleConnMedia
	db.Find(&conns)
	return conns, nil
}

func (d *DbScaleConn) InsertScaleConn(conn ScaleConnMedia) error {
	// if conn.Scale == nil {
	// 	return errors.New("No scale instance assigned to this scale connection")
	// }
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

	var conn1 ScaleConnMedia
	db.Where("scale_model=?", conn.ScaleModel).Where("scale_sn=?", conn.ScaleSn).First(&conn1)
	if conn1.ScaleModel != "" {
		log.Error("scale is already exist, with ScaleModel: %v, ScaleSn: %v", conn.ScaleModel, conn.ScaleSn)
		return errors.New("scale_id is already exist")
	}

	tx := db.Create(&conn)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (d *DbScaleConn) UpdateScaleConn(conn ScaleConnMedia) error {
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
	rowAffected := db.Model(&conn).Where("scale_sn=?", conn.ScaleSn).Updates(&conn).RowsAffected
	if rowAffected == 0 {
		return errors.New("@UpdateScaleConn failed, mybe record not existing")
	} //写成save模式不生效，又改回来了

	return nil
}

func (d *DbScaleConn) DeleteScaleConn(inConn ScaleConnMedia) error {
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
	db.Where("scale_model=?", inConn.ScaleModel).Where("scale_sn=?", inConn.ScaleSn).Delete((&inConn))
	return nil
}
