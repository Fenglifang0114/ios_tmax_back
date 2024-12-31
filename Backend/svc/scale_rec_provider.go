package svc

import (
	"path/filepath"
	"sync"

	"tmaxsrv/comm"
)

type ScaleRecProvider struct {
	mu    sync.Mutex
	myId  string
	recPb *DbScaleRec
}

type ScaleRecCheckWeigherProvider struct {
	mu    sync.Mutex
	myId  string
	recPb *DbScaleRec
}

type ScaleRecTakeInProvider struct {
	mu    sync.Mutex
	myId  string
	recPb *DbScaleRec
}

type ScaleRecTakeOutProvider struct {
	mu    sync.Mutex
	myId  string
	recPb *DbScaleRec
}

var (
	SCALE_REC_DB_FILE          = filepath.Join(comm.GetSrvDataPath(), "scalerec.db")
	SCALE_REC_DB_CHECK_FILE    = filepath.Join(comm.GetSrvDataPath(), "scalereccheck.db")
	SCALE_REC_DB_TAKE_IN_FILE  = filepath.Join(comm.GetSrvDataPath(), "scalerectakein.db")
	SCALE_REC_DB_TAKE_OUT_FILE = filepath.Join(comm.GetSrvDataPath(), "scalerectakeout.db")
)

func NewScaleRecProvider() *ScaleRecProvider {
	recPb, _ := NewDbScaleRec(SCALE_REC_DB_FILE)
	return &ScaleRecProvider{myId: "ScaleRecProvider", recPb: recPb}
}

func (p *ScaleRecProvider) GetRecsList(scale Scale, scaleModel string, scaleSn string, scaleName string) ([]ScaleRec, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	recs, err := p.recPb.GetScaleRecsList(scaleModel, scaleSn, scaleName)
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
func (p *ScaleRecProvider) DeleteAllRec(modelName string, scaleSn string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.recPb.DeleteAllScaleRec(modelName, scaleSn)
}

func NewScaleRecCheckWeigherProvider() *ScaleRecCheckWeigherProvider {
	recPb, _ := NewDbScaleRec(SCALE_REC_DB_CHECK_FILE)
	return &ScaleRecCheckWeigherProvider{myId: "ScaleRecCheckWeigherProvider", recPb: recPb}
}

func (p *ScaleRecCheckWeigherProvider) GetRecsList(scale Scale, scaleModel string, scaleSn string, scaleName string) ([]ScaleRec, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	recs, err := p.recPb.GetScaleRecsList(scaleModel, scaleSn, scaleName)
	return recs, err
}

func (p *ScaleRecCheckWeigherProvider) InsertRec(rec ScaleRec) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.recPb.InsertScaleRec(rec)
}

func (p *ScaleRecCheckWeigherProvider) DeleteRec(recId uint) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.recPb.DeleteScaleRec(recId)
}
func (p *ScaleRecCheckWeigherProvider) DeleteAllRec(modelName string, scaleSn string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.recPb.DeleteAllScaleRec(modelName, scaleSn)
}

func NewScaleRecTakeInProvider() *ScaleRecTakeInProvider {
	recPb, _ := NewDbScaleRec(SCALE_REC_DB_TAKE_IN_FILE)
	return &ScaleRecTakeInProvider{myId: "ScaleRecTakeInProvider", recPb: recPb}
}

func (p *ScaleRecTakeInProvider) GetRecsList(scale Scale, scaleModel string, scaleSn string, scaleName string) ([]ScaleRec, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	recs, err := p.recPb.GetScaleRecsList(scaleModel, scaleSn, scaleName)
	return recs, err
}

func (p *ScaleRecTakeInProvider) InsertRec(rec ScaleRec) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.recPb.InsertScaleRec(rec)
}

func (p *ScaleRecTakeInProvider) DeleteRec(recId uint) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.recPb.DeleteScaleRec(recId)
}
func (p *ScaleRecTakeInProvider) DeleteAllRec(modelName string, scaleSn string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.recPb.DeleteAllScaleRec(modelName, scaleSn)
}

func NewScaleRecTakeOutProvider() *ScaleRecTakeOutProvider {
	recPb, _ := NewDbScaleRec(SCALE_REC_DB_TAKE_OUT_FILE)
	return &ScaleRecTakeOutProvider{myId: "ScaleRecTakeOutProvider", recPb: recPb}
}

func (p *ScaleRecTakeOutProvider) GetRecsList(scale Scale, scaleModel string, scaleSn string, scaleName string) ([]ScaleRec, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	recs, err := p.recPb.GetScaleRecsList(scaleModel, scaleSn, scaleName)
	return recs, err
}

func (p *ScaleRecTakeOutProvider) InsertRec(rec ScaleRec) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.recPb.InsertScaleRec(rec)
}

func (p *ScaleRecTakeOutProvider) DeleteRec(recId uint) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.recPb.DeleteScaleRec(recId)
}

func (p *ScaleRecTakeOutProvider) DeleteAllRec(modelName string, scaleSn string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.recPb.DeleteAllScaleRec(modelName, scaleSn)
}
