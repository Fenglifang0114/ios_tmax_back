package svc

import (
	"path/filepath"
	"tmaxsrv/comm"
)

type PluRecProvider struct {
	myId  string
	recPb *DbPluRec
}

func NewPluRecProvider() *PluRecProvider {
	database := filepath.Join(comm.GetSrvDataPath(), "pluinfo.db")
	recPb, _ := NewDbPluRec(database)
	return &PluRecProvider{myId: "PluRecProvider", recPb: recPb}
}

func (p *PluRecProvider) GetRecsList() ([]PluRec, error) {
	recs, err := p.recPb.GetPluRecsList()
	return recs, err
}

func (p *PluRecProvider) GetPluPath(md5Str string) ([]PluRec, error) {
	recs, err := p.recPb.GetPluRecs(md5Str)
	return recs, err
}

func (p *PluRecProvider) InsertRec(rec PluRec) error {
	return p.recPb.InsertPluRec(rec)
}
