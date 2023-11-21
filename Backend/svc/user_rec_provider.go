package svc

import "tmaxsrv/comm"

type UserRecProvider struct {
	myId  string
	recPb *DbUserRec
}

const (
	USER_REC_DB_FILE = comm.SRV_DATA_PATH + "/" + "userrec.db"
)

func NewUserRecProvider() *UserRecProvider {
	recPb, _ := NewDbUserRec(USER_REC_DB_FILE)
	return &UserRecProvider{myId: "UserRecProvider", recPb: recPb}
}

func (p *UserRecProvider) GetRecsList() ([]UserRec, error) {
	recs, err := p.recPb.GetUserRecsList()
	return recs, err
}

func (p *UserRecProvider) InsertRec(rec UserRec) error {
	return p.recPb.InsertUserRec(rec)
}

func (p *UserRecProvider) DeleteRec(recId uint) error {
	return p.recPb.DeleteUserRec(recId)
}

func (p *UserRecProvider) ModifyRec(rec UserRec) error {
	return p.recPb.UpdateUserRec(rec)
}
