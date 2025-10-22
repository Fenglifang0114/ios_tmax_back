package svc

import (
	"path/filepath"
	"tmaxsrv/comm"
)

type ScaleConnProvider struct {
	myId   string
	connPb *DbScaleConn
}

func NewScaleConnProvider() *ScaleConnProvider {
	database := filepath.Join(comm.GetSrvDataPath(), "scaleconn.db")

	connPb, _ := NewDbScaleConn(database)
	// // add a default scale for old C51 scale, we only support one scale a time
	// var comInfo ComInfo = ComInfo{DevPath: "COM3", Baud: 115200, DataBits: 8, Parity: 0, StopBits: 0}
	// var conf MediaConf = MediaConf{}
	// conf.Type = MEDIA_COM
	// conf.MediaInfoJson, _ = json.MarshalToString(comInfo)
	// scaleConn := &ScaleConnMedia{ScaleModel: "TScale", ScaleSn: "123456", TMedia: MEDIA_COM, MediaConf: conf}

	// _ = connPb.InsertScaleConn(*scaleConn) // TODO: should deal the error of inserting default connection
	return &ScaleConnProvider{myId: "ScaleConnProvider", connPb: connPb}
}

func (p *ScaleConnProvider) GetScaleConnsList() ([]*ScaleConnMedia, error) {
	return p.connPb.GetScaleConnList()
}

func (p *ScaleConnProvider) GetScaleSrvRelList() ([]*SrvScaleRel, error) {
	return p.connPb.GetSrvScaleRelList()
}

func (p *ScaleConnProvider) UpdateSrvScaleRel(rel SrvScaleRel) error {
	return p.connPb.UpdateSrvScaleRel(rel)
}

func (p *ScaleConnProvider) DeleteSrvScaleRel(rel SrvScaleRel) error {
	return p.connPb.DeleteSrvScaleRel(rel)
}

func (p *ScaleConnProvider) DeleteSrvScaleRelByScaleId(scaleId int64) error {
	return p.connPb.DeleteSrvScaleRelByScaleId(scaleId)
}

func (p *ScaleConnProvider) InsertSrvScaleRel(rel SrvScaleRel) error {
	return p.connPb.InsertSrvScaleRel(rel)
}

func (p *ScaleConnProvider) GetScaleNameById(scaleId int64) string {
	return p.connPb.getNameById(scaleId)
}
