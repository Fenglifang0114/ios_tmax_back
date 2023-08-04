package svc

import (
	"reflect"
	"testing"
)

func TestComposeToWifiPassthData(t *testing.T) {
	type args struct {
		dataStr string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{name: "TestComposeToWifiPassthData", args: args{dataStr: "AT+CWLAP\r\n"}, want: []byte{0x5a, 0xa5, 0x00, 0x15, 0xf2, 0x02, 0x00, 0x41, 0x54, 0x2b, 0x43, 0x57, 0x4c, 0x41, 0x50, 0x0d, 0x0a, 0xe8, 0x6e, 0x8a, 0x5e, 0xa5, 0x5a}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ComposeToWifiPassthData(tt.args.dataStr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ComposeToWifiPassthData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComposeToBTPassthData(t *testing.T) {
	type args struct {
		data string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{name: "TestComposeToWifiPassthData", args: args{data: "TTM:REN-AABBCC"}, want: []byte{0x5A, 0xA5, 0x00, 0x1C, 0xF2, 0x01, 0x00, 0x54, 0x54, 0x4D, 0x3A, 0x52, 0x45, 0x4E, 0x2D, 0x41, 0x41, 0x42, 0x42, 0x43, 0x43, 0x0d, 0x0a, 0x00, 0xCB, 0x15, 0x25, 0xF6, 0xA5, 0x5A}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ComposeToBTPassthData(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ComposeToBTPassthData() = %v, want %v", got, tt.want)
			}
		})
	}
}
