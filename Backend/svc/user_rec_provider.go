package svc

import (
	"path/filepath"
	"tmaxsrv/comm"
)

type UserRecProvider struct {
	myId  string
	recPb *DbUserRec
}

func NewUserRecProvider() *UserRecProvider {
	dbFile := filepath.Join(comm.GetSrvDataPath(), "userrec.db")
	recPb, _ := NewDbUserRec(dbFile)
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
