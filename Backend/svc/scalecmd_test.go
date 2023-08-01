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
