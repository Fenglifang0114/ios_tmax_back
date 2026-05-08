package svc

import (
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type DbFlowRate struct {
	dbName string
}

// FlowRate 流速头表格
type FlowRateHeader struct {
	RecId int `gorm:"primaryKey;autoincrement;not null"`
	// 总重量
	TotalWeight float64
	//总时间
	TotalTime float64
	//平均流速
	AverageFlowRate float64
	//最小流速
	MinFlowRate float64
	//最大流速
	MaxFlowRate float64
	//流速单位
	WgtUnit string
	// 创建时间
	CreatedAt time.Time
}

// FlowRateDetail 流速明细表

type FlowRateDetail struct {
	RecId int `gorm:"primaryKey;autoincrement;not null"`
	// 头Id
	HeaderId int
	// 序号
	Id int
	// 速度
	Rate float64
	//时间
	Time float64
}

type FlowRateList struct {
	FlowRateHeader FlowRateHeader
	FlowRateDetail []FlowRateDetail
}

// 初始化数据库连接和表结构
func NewFlowRateInfo(dbName string) (*DbFlowRate, error) {
	defer func() { recover() }()
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	// 自动迁移表结构
	if err := db.AutoMigrate(
		&FlowRateHeader{},
		&FlowRateDetail{},
	); err != nil {
		return nil, err
	}

	return &DbFlowRate{dbName: dbName}, nil
}

// 新增头
func (d *DbFlowRate) AddFlowRateHeader(header *FlowRateHeader) error {
	defer func() { recover() }()
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		return err
	}
	if err := db.Create(header).Error; err != nil {
		return err
	}
	return nil
}

// 新增明细表
func (d *DbFlowRate) AddFlowRateDetail(detail *FlowRateDetail) error {
	defer func() { recover() }()
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		return err
	}
	if err := db.Create(detail).Error; err != nil {
		return err
	}
	return nil
}

// 找出头表中最大的recid
func (d *DbFlowRate) FindMaxRecId() (int, error) {
	defer func() { recover() }()
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		return 0, err

	}
	var header FlowRateHeader
	if err := db.Last(&header).Error; err != nil {
		return 0, err
	}
	return header.RecId, nil

}

// 获取所有的流速信息
// 获取所有的流速信息
func (d *DbFlowRate) GetAllFlowRateLists() ([]FlowRateList, error) {
	defer func() { recover() }()
	db, err := gorm.Open(sqlite.Open(d.dbName), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	var headers []FlowRateHeader
	// 查询所有流速头信息
	if err := db.Find(&headers).Error; err != nil {
		return nil, err
	}

	var flowRateLists []FlowRateList
	for _, header := range headers {
		var details []FlowRateDetail
		// 查询每个流速头对应的明细信息
		if err := db.Where("header_id = ?", header.RecId).Find(&details).Error; err != nil {
			return nil, err
		}
		flowRateLists = append(flowRateLists, FlowRateList{
			FlowRateHeader: header,
			FlowRateDetail: details,
		})
	}
	return flowRateLists, nil
}
