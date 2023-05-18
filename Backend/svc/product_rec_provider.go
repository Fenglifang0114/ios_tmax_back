package svc

import "tmaxsrv/comm"

type ProductRecProvider struct {
	myId  string
	recPb *DbProductRec
}

const (
	PRODUCT_REC_DB_FILE = comm.SRV_DATA_PATH + "/" + "productrec.db"
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

func (p *ProductRecProvider) DeleteRec(recId uint) error {
	return p.recPb.DeleteProductRec(recId)
}

func (p *ProductRecProvider) ModifyRec(rec ProductRec) error {
	return p.recPb.UpdateProductRec(rec)
}
