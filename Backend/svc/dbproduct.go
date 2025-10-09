package svc

import (
	"errors"
	"time"
	l "tmaxsrv/log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DbProductRec struct {
	dbName string
}

type ProductRec struct {
	RecId       int `gorm:"primaryKey;autoincrement;not null"`
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
	UpdatedAt   time.Time
	CreateBy    int  `gorm:"default:1"`
	UpdateBy    int  `gorm:"default:1"`
	Enabled     bool `gorm:"default:true"`
	CreateUser  string
	UpdateUser  string
}

func NewDbProductRec(dbName string) (*DbProductRec, error) {
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
	if err = db.AutoMigrate(&ProductRec{}); err != nil {
		l.Log.Debug("failed to migrate database of scale connection")
	}

	// // 更新Enabled为null的记录，设置UpdatedAt与创建时间一致，CreateBy和UpdateBy为1，Enabled为true
	db.Exec("UPDATE product_recs SET updated_at = created_at, create_by = 1, update_by = 1, enabled = true WHERE updated_at IS NULL")

	return &DbProductRec{dbName: dbName}, nil
}

func (d *DbProductRec) GetProductRecsList() ([]ProductRec, error) {
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
	if err = db.AutoMigrate(&ProductRec{}); err != nil {
		l.Log.Debug("failed to migrate database of product")
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
	if err = db.AutoMigrate(&ProductRec{}); err != nil {
		l.Log.Debug("failed to migrate database of scale connection")
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

func (d *DbProductRec) Insert100ProductsWithGorm(products []ProductRec) error {
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
		l.Log.Debug("failed to connect database")
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Log.Debug("failed to connect database")
	}
	if sqlDB != nil {
		defer sqlDB.Close()
	}
	//找出原来的记录
	var oldRec ProductRec
	db.Where("rec_id = ?", rec.RecId).First(&oldRec)
	if oldRec.RecId == 0 {
		return errors.New("record not found")
	}

	result := db.Model(&ProductRec{}).Where("rec_id = ?", rec.RecId).Updates(map[string]interface{}{
		"plu":          rec.Plu,
		"product_code": rec.ProductCode,
		"item_code":    rec.ItemCode,
		"category":     rec.Category,
		"product_name": rec.ProductName,
		"general_unit": rec.GeneralUnit,
		"tax_type":     rec.TaxType,
		"price":        rec.Price,
		"unit_weight":  rec.UnitWeight,
		"pretare":      rec.Pretare,
		"limit_high":   rec.LimitHigh,
		"limit_low":    rec.LimitLow,
		"update_by":    rec.UpdateBy,
		"enabled":      oldRec.Enabled,
		"updated_at":   time.Now(),
		"created_at":   oldRec.CreatedAt,
		"create_by":    oldRec.CreateBy,
		"update_user":  rec.UpdateUser,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no record updated")
	}

	return nil
}

// 获取recId最大的记录，也就是最后一个新增的记录
func (d *DbProductRec) GetLastProductRec() (ProductRec, error) {
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

	var lastRec ProductRec
	db.Order("rec_id desc").First(&lastRec)
	return lastRec, nil
}

// 批量更新数据
func (d *DbProductRec) BatchUpdateProductRec(multiRec []ProductRec) error {
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
func (d *DbProductRec) DeleteProductRec(id []int) error {
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

	// 使用Where方法指定删除条件，然后调用Delete方法删除对应记录
	result := db.Where("rec_id IN ?", id).Delete(&ProductRec{})
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (d *DbProductRec) DeleteAllProductRecs() error {
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

	// 直接使用Delete方法，不指定具体的条件，即可删除所有记录
	result := db.Exec("DELETE FROM product_recs")
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// 批量启用或者停用PLU
func (d *DbProductRec) UpdateProductRecEnabled(pluList []int, enabled bool, updateUser string) error {

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
	// 批量更新
	result := db.Model(&ProductRec{}).Where("rec_id IN ?", pluList).Updates(map[string]interface{}{

		"enabled":     enabled,
		"update_user": updateUser,
	})
	if result.Error != nil {
		return result.Error
	}
	return nil

}
