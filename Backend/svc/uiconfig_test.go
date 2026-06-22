package svc

import (
	"reflect"
	"testing"
)

func TestNewUiConfig(t *testing.T) {
	tests := []struct {
		name string
		want *UiConfig
	}{
		{name: "config", want: &UiConfig{&Config{RecMode: "auto", StableTimeToRec: "2"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewUiConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewUiConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUiConfig_GetConfig(t *testing.T) {
	c := NewUiConfig()
	tests := []struct {
		name    string
		c       *UiConfig
		want    *Config
		wantErr bool
	}{
		{name: "get config", c: c, want: &Config{RecMode: "auto", StableTimeToRec: "2"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.getConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("UiConfig.GetConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UiConfig.GetConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
