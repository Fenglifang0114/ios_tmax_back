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
	Plu         string
	ProductCode string
	ItemCode    string
	Category    string
	ProductName string
	GeneralUnit string
	TaxType     string
	Price       string
	UnitWeight  string
	Pretare     string
	LimitHigh   string
	LimitLow    string
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

func (d *DbProductRec) Insert100ProductsWithGorm(products []ProductRec) error {
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

	// 使用Create方法批量插入数据
	result := db.Create(products)
	if result.Error != nil {
		return result.Error
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
	var oldRec ProductRec
	if err := db.Where("rec_id = ?", rec.RecId).First(&oldRec).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("@UpdateProductRec failed, mybe record not existing")
		} else {
			return errors.New("@UpdateProductRec failed, mybe record not existing")
		}
	}
	db.Save(&rec)
	// rowAffected := db.Model(&rec).Where("rec_id=?", rec.RecId).Updates(&rec).RowsAffected  这个不生效
	return nil
}

// 批量更新数据
func (d *DbProductRec) BatchUpdateProductRec(multiRec []ProductRec) error {
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

	// 开启事务
	tx := db.Begin()
	for _, rec := range multiRec {
		var targetRec ProductRec
		// 先根据plu查找对应的记录，获取其rec_id
		if err := tx.Where("plu =?", rec.Plu).First(&targetRec).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 新增记录
				if err := tx.Create(&rec).Error; err != nil {
					// 回滚事务
					tx.Rollback()
					return err
				}
			} else {
				// 回滚事务
				tx.Rollback()
				return err
			}
		} else {
			// 将当前要处理的rec的rec_id设置为查找到的对应记录的rec_id，以便后续基于rec_id更新
			rec.RecId = targetRec.RecId
			// 修改记录，这里使用Save方法会根据主键（rec_id）来更新对应记录，只会更新非零值字段
			if err := tx.Save(&rec).Error; err != nil {
				// 回滚事务
				tx.Rollback()
				return err
			}
		}
	}
	// 提交事务
	return tx.Commit().Error
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

	var rec ProductRec
	rec.RecId = id
	db.Where("rec_id=?", id).Delete((&rec))
	return nil
}

func (d *DbProductRec) DeleteAllProductRecs() error {
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

	// 直接使用Delete方法，不指定具体的条件，即可删除所有记录
	result := db.Exec("DELETE FROM product_recs")
	if result.Error != nil {
		return result.Error
	}

	return nil
}
