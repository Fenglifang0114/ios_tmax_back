package svc

import (
	"reflect"
	"testing"
)

func Test_pickerFnTmaxScale(t *testing.T) {
	type args struct {
		inData  []byte
		dataLen int
	}
	tests := []struct {
		name                string
		args                args
		wantPackOffset      uint
		wantPackLen         uint
		wantShouldRemoveLen uint
		wantPack            Packet
	}{
		{name: "test tmax picker fun #1", args: args{inData: []byte{0x5a, 0xa5, 0x00, 0x14, 0xf2, 0x01, 0x00, 0x54, 0x54, 0x4d, 0x3a, 0x4f, 0x4b, 0x0d, 0x0a, 0x00, 0xce, 0x4c, 0x35, 0x01, 0xa5, 0x5a}, dataLen: 22}, wantPackOffset: 0, wantPackLen: 22, wantShouldRemoveLen: 22, wantPack: Packet{0x09, 0xf2, 0x01, 0x00, []byte{ 0x54, 0x54, 0x4d, 0x3a, 0x4f, 0x4b, 0x0d, 0x0a, 0x00}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPackOffset, gotPackLen, gotShouldRemoveLen, gotPack := pickerFnTmaxScale(tt.args.inData, tt.args.dataLen)
			if gotPackOffset != tt.wantPackOffset {
				t.Errorf("pickerFnTmaxScale() gotPackOffset = %v, want %v", gotPackOffset, tt.wantPackOffset)
			}
			if gotPackLen != tt.wantPackLen {
				t.Errorf("pickerFnTmaxScale() gotPackLen = %v, want %v", gotPackLen, tt.wantPackLen)
			}
			if gotShouldRemoveLen != tt.wantShouldRemoveLen {
				t.Errorf("pickerFnTmaxScale() gotShouldRemoveLen = %v, want %v", gotShouldRemoveLen, tt.wantShouldRemoveLen)
			}
			if !reflect.DeepEqual(gotPack, tt.wantPack) {
				t.Errorf("pickerFnTmaxScale() gotPack = %v, want %v", gotPack, tt.wantPack)
			}
		})
	}
}
