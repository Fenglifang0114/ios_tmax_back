package svc

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DbScaleRec struct {
	dbName string
}

type ScaleRec struct {
	RecId       uint   `gorm:"primaryKey;autoincrement;not null"`
	Id          string //秤记录的序号
	ScaleModel  string
	ScaleSn     string
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
	Weight      string
	WeightUnit  string
	UserNo      string
	UserName    string
	ScaleMode   string //0 =DC500 1=check Weigher 2=take in  3=take out
	ScaleName   string

	CreatedAt time.Time
}

// 将传入的字段转为数据库字段
func ConvertToDBColumnName(columnName string) string {
	var dbColumnName string
	switch columnName {
	case "Id":
		dbColumnName = "id"
	case "ScaleModel":
		dbColumnName = "scale_model"
	case "ScaleSn":
		dbColumnName = "scale_sn"
	case "PLU":
		dbColumnName = "plu"
	case "Product Code":
		dbColumnName = "product_code"
	case "Item Code":
		dbColumnName = "item_code"
	case "Category":
		dbColumnName = "category"
	case "PLU Name":
		dbColumnName = "product_name"
	case "GeneralUnit":
		dbColumnName = "general_unit"
	case "TaxType":
		dbColumnName = "tax_type"
	case "Price":
		dbColumnName = "price"
	case "UnitWeight":
		dbColumnName = "unit_weight"
	case "Pretare":
		dbColumnName = "pretare"
	case "LimitHigh":
		dbColumnName = "limit_high"
	case "LimitLow":
		dbColumnName = "limit_low"
	case "Weight":
		dbColumnName = "weight"
	case "Weight Unit":
		dbColumnName = "weight_unit"
	case "User NO.":
		dbColumnName = "user_no"
	case "User Name":
		dbColumnName = "user_name"
	case "Scale Name":
		dbColumnName = "scale_name"
	case "Date Time":
		dbColumnName = "created_at"
	default:
		// 如果传入的字段名不匹配，默认使用 "created_at"
		dbColumnName = "created_at"
	}
	return dbColumnName
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

// GetScaleWgtRecsList 从数据库中分批获取数据
func (d *DbScaleRec) GetScaleWgtRecsList(model string, sn string, name string, offset int, limit int) ([]ScaleRec, error) {
	var err error
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}
	if sqlDB != nil {
		defer sqlDB.Close()
	}
	// Migrate the schema
	if err = db.AutoMigrate(&ScaleRec{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database of scale connection: %w", err)
	}

	// 读取内容
	var recs []ScaleRec
	result := db.Where("scale_model = ?", model).
		Where("scale_sn = ?", sn).
		Offset(offset).
		Limit(limit).
		Find(&recs)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to query scale records: %w", result.Error)
	}

	return recs, nil
}

func (d *DbScaleRec) GetScaleRecsList(model string, sn string, name string, page string, pageSize string, columnName string, direction string) ([]ScaleRec, error) {
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

	// 分页参数
	pageInt := 1
	if page != "" {
		pageInt, err = strconv.Atoi(page)
		if err != nil {
			return nil, err
		}
	}

	pageSizeInt := 10
	if pageSize != "" {
		pageSizeInt, err = strconv.Atoi(pageSize)
		if err != nil {
			return nil, err
		}
	}

	offset := (pageInt - 1) * pageSizeInt

	// 读取内容
	var recs []ScaleRec
	query := db.Where("scale_model = ?", model).Where("scale_sn = ?", sn)
	dbColumnName := ConvertToDBColumnName(columnName)

	// 排序
	if dbColumnName != "" && direction != "" {
		// 将 direction 转换为 SQL 标准格式
		sqlDirection := ""
		switch direction {
		case "ascending":
			sqlDirection = "ASC"
		case "descending":
			sqlDirection = "DESC"
		default:
			// 如果 direction 不是 "ascending" 或 "descending"，默认为 "ASC"
			sqlDirection = "ASC"
		}

		// 处理按照 id 排序的情况
		if dbColumnName == "id" {
			// 使用 CAST 函数将 id 转换为整数进行排序
			query = query.Order("CAST(" + dbColumnName + " AS INTEGER) " + sqlDirection)
		} else {
			query = query.Order(dbColumnName + " " + sqlDirection)
		}

	}

	// 统计满足条件的数据总数
	var total int64
	if err := query.Model(&ScaleRec{}).Count(&total).Error; err != nil {
		return nil, err
	}

	// 计算总页数
	totalPages := (int(total) + pageSizeInt - 1) / pageSizeInt

	// 判断页码是否超出范围
	if pageInt > totalPages {
		return []ScaleRec{}, nil
	}

	// 分页
	query = query.Offset(offset).Limit(pageSizeInt)

	if err := query.Find(&recs).Error; err != nil {
		return nil, err
	}

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

func (d *DbScaleRec) DeleteAllScaleRec(scaleModel string, scaleSn string) error {
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

	// err = db.Migrator().DropTable(&ScaleRec{})
	// if err != nil {
	// 	panic("failed to drop database")
	// }

	var rec ScaleRec
	db.Where("scale_model = ? AND scale_sn = ?", scaleModel, scaleSn).Delete(&rec)

	return nil
}
