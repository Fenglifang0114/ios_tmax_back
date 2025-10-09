package svc

import (
	"path/filepath"
	"tmaxsrv/comm"
)

type ProductRecProvider struct {
	myId  string
	recPb *DbProductRec
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
	return p.recPb.Insert100ProductsWithGorm(rec)
}

func (p *ProductRecProvider) BatchModifyRec(rec []ProductRec) error {
	return p.recPb.BatchUpdateProductRec(rec)
}

func (p *ProductRecProvider) UpdateProductRecEnabled(rec []int, enabled bool, updateBy string) error {
	return p.recPb.UpdateProductRecEnabled(rec, enabled, updateBy)
}
func (p *ProductRecProvider) GetLastProductRec() (ProductRec, error) {
	return p.recPb.GetLastProductRec()
}
