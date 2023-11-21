package svc

import "tmaxsrv/comm"

type ModeSettingProvider struct {
	myId      string
	settingPb *DbModeSetting
}

const (
	MODE_SETTING_DB_FILE = comm.SRV_DATA_PATH + "/" + "modesetting.db"
)

func NewModeSettingProvider() *ModeSettingProvider {
	settingPb, _ := NewDbModeSetting(MODE_SETTING_DB_FILE)
	return &ModeSettingProvider{myId: "ModeSettingProvider", settingPb: settingPb}
}

func (p *ModeSettingProvider) GetModeSetting(scaleMode uint) ([]ModeSetting, error) {
	setting, err := p.settingPb.GetModeSetting(scaleMode)
	return setting, err
}

func (p *ModeSettingProvider) ModifyModeSe(modeSetting ModeSetting) error {
	return p.settingPb.UpdateModeSetting(modeSetting)
}
