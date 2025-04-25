package svc

import (
	"path/filepath"
	"tmaxsrv/comm"
)

type FlowRateProvider struct {
	myId   string
	infoPb *DbFlowRate
}

func NewFlowRateProvider() *FlowRateProvider {
	database := filepath.Join(comm.GetSrvDataPath(), "flowrate.db")
	infoPb, _ := NewFlowRateInfo(database)
	return &FlowRateProvider{myId: "FlowrateProvider", infoPb: infoPb}
}

// 获取流速列表
func (p *FlowRateProvider) GetFlowRateList() ([]FlowRateList, error) {
	recs, err := p.infoPb.GetAllFlowRateLists()
	return recs, err
}

// 新增流速头
func (p *FlowRateProvider) InsertFlowRateHeader(header *FlowRateHeader) error {
	return p.infoPb.AddFlowRateHeader(header)
}

// 新增流速明细表
func (p *FlowRateProvider) InsertFlowRateDetail(detail *FlowRateDetail) error {
	return p.infoPb.AddFlowRateDetail(detail)
}
