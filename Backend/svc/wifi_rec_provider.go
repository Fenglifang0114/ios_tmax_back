package svc

import "tmaxsrv/comm"

type WifiRecProvider struct {
	myId  string
	recPb *DbWifiRec
}

const (
	WIFI_REC_DB_FILE = comm.SRV_DATA_PATH + "/" + "wifirec.db"
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
