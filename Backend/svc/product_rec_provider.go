package svc

import (
	"path/filepath"
	"sync"
	"tmaxsrv/comm"
)

type ProductRecProvider struct {
	myId  string
	recPb *DbProductRec
	mu    sync.Mutex
}

var (
	PRODUCT_REC_DB_FILE = filepath.Join(comm.GetSrvDataPath(), "productrec.db")
)

func NewProductRecProvider() *ProductRecProvider {
	recPb, _ := NewDbProductRec(PRODUCT_REC_DB_FILE)
	return &ProductRecProvider{myId: "ProductRecProvider", recPb: recPb}
}

func (p *ProductRecProvider) GetRecsList() ([]ProductRec, error) {
	recs, err := p.recPb.GetProductRecsList()
	return recs, err
}

// CheckPluExist 检查PLU是否存在
func (p *ProductRecProvider) CheckPluExist(id int, plu string) (bool, error) {
	return p.recPb.CheckPluExist(id, plu)
}

// 分页获取PLU
func (p *ProductRecProvider) GetPluByPage(page, pageSize int, fieldName, direction string, search ProductQuery) ([]ProductRec, int, error) {
	recs, total, err := p.recPb.GetPluByPage(page, pageSize, fieldName, direction, search)
	return recs, total, err
}

// 导出产品列表
func (p *ProductRecProvider) GetExportProductList(searchPlu ProductQuery) ([]ProductRec, error) {
	return p.recPb.GetExportProductList(searchPlu)
}

func (p *ProductRecProvider) InsertRec(rec ProductRec) error {
	return p.recPb.InsertProductRec(rec)
}

func (p *ProductRecProvider) DeleteRec(recId []int) error {

	return p.recPb.DeleteProductRec(recId)
}

func (p *ProductRecProvider) DeleteAllRec() error {
	return p.recPb.DeleteAllProductRecs()
}

func (p *ProductRecProvider) ModifyRec(rec ProductRec) error {
	return p.recPb.UpdateProductRec(rec)
}

func (p *ProductRecProvider) Insert100Rec(rec []ProductRec) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.recPb.Insert100ProductsWithGorm(rec)
}

func (p *ProductRecProvider) BatchModifyRec(rec []ProductRec) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.recPb.BatchUpdateProductRec(rec)
}

func (p *ProductRecProvider) UpdateProductRecEnabled(rec []int, enabled bool, updateBy string) error {
	return p.recPb.UpdateProductRecEnabled(rec, enabled, updateBy)
}
func (p *ProductRecProvider) GetLastProductRec() (ProductRec, error) {
	return p.recPb.GetLastProductRec()
}

// 获取PLU设置
func (p *ProductRecProvider) GetPluSetting() (map[int]string, error) {
	return p.recPb.GetPluSetting()
}

// 设置PLU显示字段
func (p *ProductRecProvider) SetPluSetting(fieldDisplaySetting []string, updateUser string) error {
	return p.recPb.SetPluSetting(fieldDisplaySetting, updateUser)
}
