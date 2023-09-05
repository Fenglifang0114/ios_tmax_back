package picker

import (
	"reflect"
	"testing"

	"tmaxsrv/comm"
)

func Test_verifyTailT2200(t *testing.T) {
	type args struct {
		buf     []byte
		headPos int
		len     int
	}
	tests := []struct {
		name  string
		args  args
		want  int
		want1 State
	}{
		{name: "Test_verifyTailT2200 #1", args: args{buf: []byte{0x5a, 0xa5, 0x00, 0x01, 0x06, 0x7f, 0xfb, 0xf1, 0x6e, 0xa5, 0x5a}, headPos: 0, len: 11}, want: 9, want1: FOUND},
		{name: "Test_verifyTailT2200 #2", args: args{buf: []byte{0x5a, 0xa5, 0x00, 0x01, 0x06, 0x7f, 0xfb, 0xf1, 0x6e, 0xa4, 0x5a}, headPos: 0, len: 11}, want: -1, want1: NOT_FOUND},
		{name: "Test_verifyTailT2200 #3", args: args{buf: []byte{0x5a, 0xa5, 0x00, 0x02, 0x06, 0x7f, 0xfb, 0xf1, 0x6e, 0xa5, 0x5a}, headPos: 0, len: 11}, want: -1, want1: NOT_ENOUGH_DATA},
		{name: "Test_verifyTailT2200 #4", args: args{buf: []byte{0x5a, 0xa5, 0x00, 0x02, 0x06, 0x7f, 0xfb, 0xf1, 0x6e, 0xa5, 0x5a, 0x55}, headPos: 0, len: 12}, want: -1, want1: NOT_FOUND},
		{name: "Test_verifyTailT2200 #4", args: args{buf: []byte{0x5a, 0xa5, 0x00, 0x02, 0x06, 0x00, 0x7f, 0xfb, 0xf1, 0x6e, 0xa5, 0x5a, 0x55}, headPos: 0, len: 13}, want: 10, want1: FOUND},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := verifyTailT2200(tt.args.buf, tt.args.headPos, tt.args.len)
			if got != tt.want {
				t.Errorf("verifyTailT2200() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("verifyTailT2200() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_verifyCrcT2200(t *testing.T) {
	type args struct {
		buf     []byte
		headPos int
		tailPos int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "Test_verifyCrcT2200 #1", args: args{buf: []byte{0x77, 0x55, 0x5a, 0xa5, 0x00, 0x01, 0x06, 0x7f, 0xfb, 0xf1, 0x6e, 0xa5, 0x5a, 0x5a, 0xa5}, headPos: 2, tailPos: 11}, want: true},
		{name: "Test_verifyCrcT2200 #2", args: args{buf: []byte{0x77, 0x55, 0x5a, 0xa5, 0x00, 0x01, 0x06, 0x6f, 0xfb, 0xf1, 0x6e, 0xa5, 0x5a, 0x5a, 0xa5}, headPos: 2, tailPos: 11}, want: false},
		{name: "Test_verifyCrcT2200 #3", args: args{buf: []byte{0x77, 0x55, 0x5a, 0xa5, 0x00, 0x01, 0x06, 0x6f, 0xfb, 0xf1}, headPos: 2, tailPos: 11}, want: false},
		{name: "Test_verifyCrcT2200 #4", args: args{buf: []byte{0x77, 0x55, 0x5a, 0xa5, 0x00, 0x01, 0x15, 0x7f, 0xfb, 0xf1, 0x6e, 0xa5, 0x5a, 0x5a, 0xa5}, headPos: 2, tailPos: 11}, want: false},
		{name: "Test_verifyCrcT2200 #5", args: args{buf: []byte{0x77, 0x55, 0x5a, 0xa5, 0x00, 0x01, 0x15, 0x3E, 0xA9, 0x0C, 0xC7, 0xa5, 0x5a, 0x5a, 0xa5}, headPos: 2, tailPos: 11}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := verifyCrcT2200(tt.args.buf, tt.args.headPos, tt.args.tailPos); got != tt.want {
				t.Errorf("verifyCrcT2200() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPickerFnT2200(t *testing.T) {
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
		wantPack            comm.Packet
	}{
		{name: "TestPickerFnT2200 #1", args: args{inData: []byte{0x77, 0x55, 0x5a, 0xa5, 0x00, 0x01, 0x06, 0x7f, 0xfb, 0xf1, 0x6e, 0xa5, 0x5a, 0x5a, 0xa5}, dataLen: 14}, wantPackOffset: 2, wantPackLen: 11, wantShouldRemoveLen: 13, wantPack: comm.Packet{PayloadLen: 1, Payload: []byte{0x06}}},
		{name: "TestPickerFnT2200 #2", args: args{inData: []byte{0x77, 0x55, 0x5a, 0xa5, 0x00, 0x01, 0x15, 0x7f, 0xfb, 0xf1, 0x6e, 0xa5, 0x5a, 0x5a, 0xa5}, dataLen: 15}, wantPackOffset: 0, wantPackLen: 0, wantShouldRemoveLen: 13, wantPack: comm.Packet{}},
		{name: "TestPickerFnT2200 #3", args: args{inData: []byte{0x77, 0x55, 0x5a, 0xa5, 0x00, 0x01, 0x15, 0x3E, 0xA9, 0x0C, 0xC7, 0xa5, 0x5a, 0x5a, 0xa5}, dataLen: 15}, wantPackOffset: 2, wantPackLen: 11, wantShouldRemoveLen: 13, wantPack: comm.Packet{PayloadLen: 1, Payload: []byte{0x15}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPackOffset, gotPackLen, gotShouldRemoveLen, gotPack := PickerFnT2200(tt.args.inData, tt.args.dataLen)
			if gotPackOffset != tt.wantPackOffset {
				t.Errorf("PickerFnT2200() gotPackOffset = %v, want %v", gotPackOffset, tt.wantPackOffset)
			}
			if gotPackLen != tt.wantPackLen {
				t.Errorf("PickerFnT2200() gotPackLen = %v, want %v", gotPackLen, tt.wantPackLen)
			}
			if gotShouldRemoveLen != tt.wantShouldRemoveLen {
				t.Errorf("PickerFnT2200() gotShouldRemoveLen = %v, want %v", gotShouldRemoveLen, tt.wantShouldRemoveLen)
			}
			if !reflect.DeepEqual(gotPack, tt.wantPack) {
				t.Errorf("PickerFnT2200() gotPack = %v, want %v", gotPack, tt.wantPack)
			}
		})
	}
}
