package svc

import "tmaxsrv/comm"

type ScaleConnProvider struct {
	myId   string
	connPb *DbScaleConn
}

const (
	SCALE_CONN_DB_FILE = comm.SRV_DATA_PATH + "/" + "scaleconn.db"
)

func NewScaleConnProvider() *ScaleConnProvider {
	connPb, _ := NewDbScaleConn(SCALE_CONN_DB_FILE)
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
