package svc

import (
	"sync"

	"tmaxsrv/comm"
)

type ScaleRecProvider struct {
	mu    sync.Mutex
	myId  string
	recPb *DbScaleRec
}

const (
	SCALE_REC_DB_FILE = comm.SRV_DATA_PATH + "/" + "scalerec.db"
)

func NewScaleRecProvider() *ScaleRecProvider {
	recPb, _ := NewDbScaleRec(SCALE_REC_DB_FILE)
	return &ScaleRecProvider{myId: "ScaleRecProvider", recPb: recPb}
}

func (p *ScaleRecProvider) GetRecsList(scale Scale) ([]ScaleRec, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	recs, err := p.recPb.GetScaleRecsList(scale.Model, scale.Sn)
	return recs, err
}

func (p *ScaleRecProvider) InsertRec(rec ScaleRec) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.recPb.InsertScaleRec(rec)
}

func (p *ScaleRecProvider) DeleteRec(recId uint) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.recPb.DeleteScaleRec(recId)
}
