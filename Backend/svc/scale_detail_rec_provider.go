package svc

import (
	"path/filepath"
	"sync"
	"tmaxsrv/comm"
)

type DetailRecProvider struct {
	mu       sync.Mutex
	myId     string
	detailPb *DbDetailRec
}

var (
	DETAIL_REC_DB_FILE = filepath.Join(comm.GetSrvDataPath(), "detailrec.db")
)

func NewDetailRecProvider() *DetailRecProvider {
	detailPb, _ := NewDbDetailRec(DETAIL_REC_DB_FILE)
	return &DetailRecProvider{myId: "DetailRecProvider", detailPb: detailPb}
}

func (p *DetailRecProvider) GetRecsList() ([]DetailList, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	recs, err := p.detailPb.GetDetailRecsList()
	return recs, err
}

func (p *DetailRecProvider) InsertTotalRec(rec DetailTotal) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.detailPb.InsertDetailTotal(rec)
}

func (p *DetailRecProvider) InsertDetailRec(rec DetailRec) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.detailPb.InsertDetailRec(rec)
}

func (p *DetailRecProvider) DeleteRec(recId uint) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.detailPb.DeleteDetailRec(recId)
}
func (p *DetailRecProvider) DeleteAllRec(modelName string, scaleSn string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.detailPb.DeleteAllDetailRec(modelName, scaleSn)
}
