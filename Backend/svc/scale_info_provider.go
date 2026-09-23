package svc

import (
	"path/filepath"
	"sync"

	"tmaxsrv/comm"
)

type ScaleInfosProvider struct {
	mu     sync.Mutex
	myId   string
	infoPb *DbScaleInfos
}

func NewScaleInfosProvider() *ScaleInfosProvider {
	dbFile := filepath.Join(comm.GetSrvDataPath(), "scaleinfo.db")
	infoPb, _ := NewDbScaleInfos(dbFile)
	return &ScaleInfosProvider{myId: "ScaleInfosProvider", infoPb: infoPb}
}

func (p *ScaleInfosProvider) InsertRec(info ScaleInfos) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.infoPb.InsertScaleInfos(info)
}

func (p *ScaleInfosProvider) DeleteRec(recId uint) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.infoPb.DeleteScaleInfos(recId)
}
func (p *ScaleInfosProvider) DeleteAllInfos() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.infoPb.DeleteAllScaleInfos()
}
