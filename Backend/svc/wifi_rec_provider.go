package svc

import (
	"path/filepath"
	"tmaxsrv/comm"
)

type WifiRecProvider struct {
	myId  string
	recPb *DbWifiRec
}

func NewWifiRecProvider() *WifiRecProvider {
	dbFile := filepath.Join(comm.GetSrvDataPath(), "wifirec.db")
	recPb, _ := NewDbWifiRec(dbFile)
	return &WifiRecProvider{myId: "WifiRecProvider", recPb: recPb}
}

func (p *WifiRecProvider) GetRecsList() ([]WifiRec, error) {
	recs, err := p.recPb.GetWifiRecsList()
	return recs, err
}

func (p *WifiRecProvider) InsertRec(rec WifiRec) error {
	return p.recPb.InsertWifiRec(rec)
}

func (p *WifiRecProvider) DeleteRec(recId uint) error {
	return p.recPb.DeleteWifiRec(recId)
}

func (p *WifiRecProvider) ModifyRec(rec WifiRec) error {
	return p.recPb.UpdateWifiRec(rec)
}
