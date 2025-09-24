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

// 新增原料类别列表
func (p *FormulaRecProvider) InsertRawTypeList(rec []string) error {
	return p.infoPb.CreateRawCategoryList(rec)
}

// 新增配方类别列表
func (p *FormulaRecProvider) InsertFormulaTypeList(rec []string) error {
	return p.infoPb.CreateFormulaCategoryList(rec)
}

// 新增原料信息
func (p *FormulaRecProvider) InsertRawInfo(rec RawMaterial) error {
	return p.infoPb.CreateRawMaterial(rec)
}

// 新增原料信息列表
func (p *FormulaRecProvider) InsertRawInfoList(rec []RawMaterial) error {
	return p.infoPb.CreateRawMaterialList(rec)
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

// 删除所有原料信息
func (p *FormulaRecProvider) DeleteAllRawInfo(recIds []int) error {
	return p.infoPb.DeleteAllRawMaterials(recIds)
}

// 获取配方信息头ID
func (p *FormulaRecProvider) GetMaxFormulaRecId() (int, error) {
	return p.infoPb.GetMaxFormulaRecId()
}

// 获取配方信息头key
func (p *FormulaRecProvider) GetMaxFormulaRecKey() (int, error) {
	return p.infoPb.GetMaxFormulaRecKey()
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

// 删除所有配方
func (p *FormulaRecProvider) DeleteAllFormulas(rec_ids []int) error {
	return p.infoPb.DeleteAllFormulaByRecId(rec_ids)
}

// 修改配方
func (p *FormulaRecProvider) UpdateFormula(header FormulaHeader, details []FormulaDetail) error {
	return p.infoPb.UpdateFormula(header, details)

}

// 删除原料类型
func (p *FormulaRecProvider) DeleteRawType(recId string) error {
	return p.infoPb.DeleteRawMaterialCategory(recId)
}

// 修改原料类型
func (p *FormulaRecProvider) UpdateRawType(rec RawMaterialCategory) error {
	return p.infoPb.UpdateRawMaterialCategory(rec)
}

// 新增配方
func (p *FormulaRecProvider) InsertFmaInfoList(rec []FmaDataImportInfo) error {
	return p.infoPb.InsertFormulaList(rec)
}

// 删除配方类型
func (p *FormulaRecProvider) DeleteFmaType(name string) error {
	return p.infoPb.DeleteFormulaCategory(name)

}

// 修改配方类型
func (p *FormulaRecProvider) UpdateFmaType(rec FormulaCategory) error {
	return p.infoPb.UpdateFormulaCategory(rec)
}

// 获取配方秤中的自动下一步设置
func (p *FormulaRecProvider) GetSetAutoNext() (*SetAutoNext, error) {
	return p.infoPb.GetSetAutoNext()
}

// 更新配方秤中的自动下一步设置
func (p *FormulaRecProvider) UpdateSetAutoNext(rec SetAutoNext) error {
	return p.infoPb.UpdateSetAutoNext(rec.AutoNext, rec.StableTime, rec.AutoTare)
}

// 创建暂存配方称重记录
func (p *FormulaRecProvider) CreateDraftFmaWgtRecHeader(rec DrafFmaWgtRecHeader) error {
	return p.infoPb.CreateDraftFmaWgtRecHeader(rec)
}

// 创建暂存配方称重记录体
func (p *FormulaRecProvider) CreateDraftFmaWgtRecDetail(rec DrafFmaWgtRecDetail) error {
	return p.infoPb.CreateDraftFmaWgtRecDetail(rec)
}

// 更新暂存配方称重记录
func (p *FormulaRecProvider) UpdateDraftFmaWgtRec(rec DrafFmaWgtRecInfo) error {
	return p.infoPb.UpdateDraftFmaWgtRec(rec)
}

// 获取暂存配方称重记录
func (p *FormulaRecProvider) GetDraftFmaWgtRec() ([]DrafFmaWgtRecInfo, error) {
	return p.infoPb.GetAllDraftFmaWgtRecLists()

}

// 删除暂存配方称重记录
func (p *FormulaRecProvider) DeleteDraftFmaWgtRec(orderId string) error {
	return p.infoPb.DeleteDraftFmaWgtRec(orderId)
}

// 删除所有暂存配方称重记录
func (p *FormulaRecProvider) DeleteAllDraftFmaWgtRec(orderIds []string) error {
	return p.infoPb.DeleteAllDraftFmaWgtRec(orderIds)
}

// 检查配方材料里面是否使用了这个秤
func (p *FormulaRecProvider) CheckFormulaRawData(scaleId int) (bool, error) {
	return p.infoPb.CheckFormulaRawData(scaleId)
}
