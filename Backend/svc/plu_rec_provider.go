package svc

import "tmaxsrv/comm"

type PluRecProvider struct {
	myId  string
	recPb *DbPluRec
}

const (
	PLU_REC_DB_FILE = comm.SRV_DATA_PATH + "/" + "pluinfo.db"
)

func NewPluRecProvider() *PluRecProvider {
	recPb, _ := NewDbPluRec(PLU_REC_DB_FILE)
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
