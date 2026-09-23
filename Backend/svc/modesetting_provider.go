package svc

import (
	"path/filepath"
	"tmaxsrv/comm"
)

type ModeSettingProvider struct {
	myId      string
	settingPb *DbModeSetting
}

func NewModeSettingProvider() *ModeSettingProvider {
	dbFile := filepath.Join(comm.GetSrvDataPath(), "modesetting.db")
	settingPb, _ := NewDbModeSetting(dbFile)
	return &ModeSettingProvider{myId: "ModeSettingProvider", settingPb: settingPb}
}

func (p *ModeSettingProvider) GetModeSetting(scaleMode uint) ([]ModeSetting, error) {
	setting, err := p.settingPb.GetModeSetting(scaleMode)
	return setting, err
}

func (p *ModeSettingProvider) ModifyModeSe(modeSetting ModeSetting) error {
	return p.settingPb.UpdateModeSetting(modeSetting)
}
