package svc

import (
	"path/filepath"
	"tmaxsrv/comm"
)

type FormulaRecProvider struct {
	myId   string
	infoPb *DbFormulaInfo
}

func NewFormulaRecProvider() *FormulaRecProvider {
	database := filepath.Join(comm.GetSrvDataPath(), "formulainfo.db")
	infoPb, _ := NewFormulaInfo(database)
	return &FormulaRecProvider{myId: "FormulaRecProvider", infoPb: infoPb}
}

// 获取所有配方列表
func (p *FormulaRecProvider) GetFormulaRecsList() ([]FormulaList, error) {
	recs, err := p.infoPb.GetAllFormulaLists()
	return recs, err
}

// 获取所有重量模式配方记录列表
func (p *FormulaRecProvider) GetPluPath() ([]FormulaWgtRecList, error) {
	recs, err := p.infoPb.GetAllFormulaWgtRecLists()
	return recs, err
}

// 新增原料类别
func (p *FormulaRecProvider) InsertRawType(rec RawMaterialCategory) error {
	return p.infoPb.CreateRawMaterialCategory(rec)
}

// 新增原料信息
func (p *FormulaRecProvider) InsertRawInfo(rec RawMaterial) error {
	return p.infoPb.CreateRawMaterial(rec)
}

// 新增配方类别
func (p *FormulaRecProvider) InsertFormulaType(rec FormulaCategory) error {
	return p.infoPb.CreateFormulaCategory(rec)
}

// 获取原料类别列表
func (p *FormulaRecProvider) GetRawTypeList() ([]RawMaterialCategory, error) {
	recs, err := p.infoPb.GetAllRawMaterialCategories()
	return recs, err
}

// 获取配方类别列表
func (p *FormulaRecProvider) GetFormulaTypeList() ([]FormulaCategory, error) {
	recs, err := p.infoPb.GetAllFormulaCategories()
	return recs, err
}

// 获取原料列表
func (p *FormulaRecProvider) GetRawDataList() ([]RawMaterialWithTypeName, error) {
	recs, err := p.infoPb.GetAllRawMaterialsWithTypeName()
	return recs, err
}

// 修改原料信息
func (p *FormulaRecProvider) UpdateRawInfo(rec RawMaterial) error {
	return p.infoPb.UpdateRawMaterial(rec)
}

// 删除原料信息
func (p *FormulaRecProvider) DeleteRawInfo(recId int) error {
	return p.infoPb.DeleteRawMaterial(recId)
}

// 新增配方信息头
func (p *FormulaRecProvider) InsertFormulaHeader(rec FormulaHeader) error {
	return p.infoPb.CreateFormulaHeader(rec)
}

// 新增配方信息体
func (p *FormulaRecProvider) InsertFormulaBody(rec FormulaDetail) error {
	return p.infoPb.CreateFormulaDetail(rec)
}

// 获取配方信息列表
func (p *FormulaRecProvider) GetFormulaDataList() ([]FormulaList, error) {

	return p.infoPb.GetAllFormulaLists()
}

// 根据配方编号获取配方
func (p *FormulaRecProvider) GetFormulaListByFormulaID(formulaID string) (FormulaList, error) {
	recs, err := p.infoPb.GetFormulaListByFormulaID(formulaID)
	return recs, err
}

// 新增配方称重记录头
func (p *FormulaRecProvider) InsertFormulaWgtHeader(rec FormulaWgtRecHeader) error {
	return p.infoPb.CreateFormulaWgtRecHeader(rec)
}

// 新增配方称重记录体
func (p *FormulaRecProvider) InsertFormulaWgtBody(rec FormulaWgtRecDetail) error {
	return p.infoPb.CreateFormulaWgtRecDetail(rec)
}

// 获取配方称重记录列表
func (p *FormulaRecProvider) GetFormulaWgtRecList() ([]FormulaWgtRecList, error) {
	recs, err := p.infoPb.GetAllFormulaWgtRecLists()
	return recs, err
}

// 删除配方
func (p *FormulaRecProvider) DeleteFormula(rec_id int) error {
	return p.infoPb.DeleteFormulaByRecId(rec_id)
}
