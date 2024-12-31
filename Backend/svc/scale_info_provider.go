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

var (
	SCALE_INFO_DB_FILE = filepath.Join(comm.GetSrvDataPath(), "scaleinfo.db")
)

func NewScaleInfosProvider() *ScaleInfosProvider {
	infoPb, _ := NewDbScaleInfos(SCALE_INFO_DB_FILE)
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
