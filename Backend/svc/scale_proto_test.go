package svc

import (
	"reflect"
	"testing"
)

func TestModifyBTNameCmd(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// TODO: Add test cases. 5A A5 00 1C F2 01 00 54 54 4D 3A 52 45 4E 2D 41 41 42 42 43 43 0d 0a 00 CB 15 25 F6 A5 5A
		{name: "Modify BT Cmd Test", args: args{name: "AABBCC"}, want: []byte{0x5A, 0xA5, 0x00, 0x1C, 0xF2, 0x01, 0x00, 0x54, 0x54, 0x4D, 0x3A, 0x52, 0x45, 0x4E, 0x2D, 0x41, 0x41, 0x42, 0x42, 0x43, 0x43, 0x0d, 0x0a, 0x00, 0xCB, 0x15, 0x25, 0xF6, 0xA5, 0x5A}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ModifyBTNameCmd(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ModifyBTNameCmd() = %v, want %v", got, tt.want)
			}
		})
	}
}
