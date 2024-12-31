package svc

import (
	"path/filepath"
	"tmaxsrv/comm"
)

type WifiRecProvider struct {
	myId  string
	recPb *DbWifiRec
}

var (
	WIFI_REC_DB_FILE = filepath.Join(comm.GetSrvDataPath(), "wifirec.db")
)

func NewWifiRecProvider() *WifiRecProvider {
	recPb, _ := NewDbWifiRec(WIFI_REC_DB_FILE)
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
