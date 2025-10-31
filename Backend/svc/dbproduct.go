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

type FieldDisplaySetting struct {
	Id        int `gorm:"primaryKey;autoincrement;not null"`
	FieldName string
	Sequence  int
	UpdatedAt time.Time
	UserName  string
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
	if err = db.AutoMigrate(&ProductRec{}, &FieldDisplaySetting{}); err != nil {
		l.Log.Debug("failed to migrate database of scale connection")
	}

	// // 更新Enabled为null的记录，设置UpdatedAt与创建时间一致，CreateBy和UpdateBy为1，Enabled为true
	db.Exec("UPDATE product_recs SET updated_at = created_at, create_by = 1, update_by = 1, enabled = true WHERE updated_at IS NULL")

	return &DbProductRec{dbName: dbName}, nil
}

// CheckPluExist 检查PLU是否存在
func (d *DbProductRec) CheckPluExist(id int, plu string) (bool, error) {
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

	//查找Plu信息
	var recs []ProductRec
	query := db.Model(&ProductRec{}).Where("plu = ?", plu)
	query.Find(&recs)
	if id == 0 {
		return len(recs) > 0, nil
	}

	if len(recs) > 0 {
		return recs[0].RecId != id, nil
	}

	return false, nil

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

// 获取分页数据
func (d *DbProductRec) GetPluByPage(page, pageSize int, fieldName, direction string, search ProductQuery) ([]ProductRec, int, error) {
	// fieldName是排序字段，direction是排序方向asc/desc
	// ProductQuery是查询条件,如果为空,则不进行过滤
	var err error
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		l.Log.Debug("failed to connect database")
		return nil, 0, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Log.Debug("failed to connect database")
		return nil, 0, err
	}
	if sqlDB != nil {
		defer sqlDB.Close()
	}
	// Migrate the schema
	if err = db.AutoMigrate(&ProductRec{}); err != nil {
		l.Log.Debug("failed to migrate database of product")
		return nil, 0, err
	}

	//查询总数
	var total int64
	db.Model(&ProductRec{}).Count(&total)

	// 构建查询
	query := db.Model(&ProductRec{})

	// 添加搜索条件
	if search.Plu != "" {
		query = query.Where("plu LIKE ?", "%"+search.Plu+"%")
	}
	if search.PluName != "" {
		query = query.Where("product_name LIKE ?", "%"+search.PluName+"%")
	}
	if search.Category != "" {
		query = query.Where("category LIKE ?", "%"+search.Category+"%")
	}
	// Enabled 条件 - 只有当明确查询启用或禁用时才添加条件

	if search.SetEnabled {
		if search.Enabled {
			query = query.Where("enabled = ?", true)
		} else if !search.Enabled {
			query = query.Where("enabled = ?", false)
		}

	}

	// 添加排序
	if fieldName != "" {
		// 验证排序方向
		orderDirection := "ASC"
		if direction == "desc" {
			orderDirection = "DESC"
		}
		//如果字段是plu,转为数字排序
		switch fieldName {
		case "plu":
			query = query.Order("CAST(plu AS INTEGER) " + orderDirection)
		case "price":
			query = query.Order("CAST(price AS REAL) " + orderDirection)
		case "limit_high":
			query = query.Order("CAST(limit_high AS REAL) " + orderDirection)
		case "limit_low":
			query = query.Order("CAST(limit_low AS REAL) " + orderDirection)
		case "unit_weight":
			query = query.Order("CAST(unit_weight AS REAL) " + orderDirection)
		case "pretare":
			query = query.Order("CAST(pretare AS REAL) " + orderDirection)
		case "tax_type":
			query = query.Order("CAST(tax_type AS INTEGER) " + orderDirection)
		default:
			query = query.Order(fieldName + " " + orderDirection)
		}

	} else {
		// 如果没有指定排序字段，默认按plu从小到大排序
		// query = query.Order("plu ASC")
	}

	// 读取内容
	var recs []ProductRec
	result := query.Limit(pageSize).Offset((page - 1) * pageSize).Find(&recs)
	if result.Error != nil {
		l.Log.Debug("failed to query products")
		return nil, 0, result.Error
	}

	return recs, int(total), nil
}

// 导出产品列表
func (d *DbProductRec) GetExportProductList(search ProductQuery) ([]ProductRec, error) {
	var err error
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		l.Log.Debug("failed to connect database")
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		l.Log.Debug("failed to connect database")
		return nil, err
	}
	if sqlDB != nil {
		defer sqlDB.Close()
	}
	// Migrate the schema
	if err = db.AutoMigrate(&ProductRec{}); err != nil {
		l.Log.Debug("failed to migrate database of product")
		return nil, err
	}

	//查询总数
	var total int64
	db.Model(&ProductRec{}).Count(&total)

	// 构建查询
	query := db.Model(&ProductRec{})

	// 添加搜索条件
	if search.Plu != "" {
		query = query.Where("plu LIKE ?", "%"+search.Plu+"%")
	}
	if search.PluName != "" {
		query = query.Where("product_name LIKE ?", "%"+search.PluName+"%")
	}
	if search.Category != "" {
		query = query.Where("category LIKE ?", "%"+search.Category+"%")
	}
	// Enabled 条件 - 只有当明确查询启用或禁用时才添加条件

	if search.SetEnabled {
		if search.Enabled {
			query = query.Where("enabled = ?", true)
		} else if !search.Enabled {
			query = query.Where("enabled = ?", false)
		}

	}

	// 读取内容
	var recs []ProductRec
	result := query.Find(&recs)
	if result.Error != nil {
		l.Log.Debug("failed to query products")
		return nil, result.Error
	}
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

	var pluList []string
	for _, p := range products {
		pluList = append(pluList, p.Plu)
	}

	// 明确指定模型
	var existPluList []string
	err = db.Model(&ProductRec{}).Where("plu IN ?", pluList).Pluck("plu", &existPluList).Error
	if err != nil {

		return err
	}

	existPluSet := make(map[string]bool)
	for _, plu := range existPluList {
		existPluSet[plu] = true
	}

	// 过滤掉已存在的记录
	filteredProducts := make([]ProductRec, 0, len(products))
	for _, p := range products {
		if !existPluSet[p.Plu] {
			filteredProducts = append(filteredProducts, p)
		}
	}

	if len(filteredProducts) == 0 {
		return nil
	}

	result := db.Create(filteredProducts)
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

func (d *DbProductRec) GetPluSetting() (map[int]string, error) {

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
	var results []FieldDisplaySetting
	err = db.Model(&FieldDisplaySetting{}).Order("sequence ASC").Scan(&results).Error
	if err != nil {
		return nil, err
	}

	fieldDisplaySetting := make(map[int]string)
	for _, r := range results {
		fieldDisplaySetting[r.Sequence] = r.FieldName
	}

	return fieldDisplaySetting, nil
}

// 设置字段显示
func (d *DbProductRec) SetPluSetting(fieldDisplaySetting []string, updateUser string) error {

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

	//删除所有记录
	result := db.Exec("DELETE FROM field_display_settings")
	if result.Error != nil {
		return result.Error
	}
	sequence := 1
	for _, r := range fieldDisplaySetting {
		if r == "plu" || r == "productName" {
			continue
		}
		db.Create(&FieldDisplaySetting{
			Sequence:  sequence,
			FieldName: r,
			UpdatedAt: time.Now(),
			UserName:  updateUser,
		})
		sequence++
	}

	return nil

}
