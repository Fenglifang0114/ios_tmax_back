package svc

import (
	"reflect"
	"testing"
)

func TestGetString(t *testing.T) {
	type args struct {
		id CmdID
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "TestGetString", args: args{id: 0x05f0}, want: "model_sn", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetString(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewRingBuffers(t *testing.T) {
	type args struct {
		cmds []string
	}
	tests := []struct {
		name string
		args args
		want RingBuffers
	}{
		{name: "TestNewRingBuffers", args: args{cmds: []string{"test1", "test2", "test3"}}, want: RingBuffers{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRingBuffers(tt.args.cmds...); len(got) != 3 {
				t.Errorf("NewRingBuffers() = %v, length not correct", got)
			}
		})
	}
}

func TestRingBuffers_Write(t *testing.T) {
	cmdStrings := []string{"test1", "test2", "test3"}
	ringBufs := NewRingBuffers(cmdStrings...)
	type args struct {
		cmd  string
		data []byte
	}
	tests := []struct {
		name    string
		r       RingBuffers
		args    args
		wantErr bool
	}{
		{name: "TestRingBuffers_Write #1", r: ringBufs, args: args{cmd: "test1", data: []byte{0x01, 0x02, 0x03}}, wantErr: false},
		{name: "TestRingBuffers_Write #2", r: ringBufs, args: args{cmd: "test2", data: []byte{0x01, 0x02, 0x03}}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.Write(tt.args.cmd, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("RingBuffers.Write() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRingBuffers_Read(t *testing.T) {
	cmdStrings := []string{"test1", "test2", "test3"}
	ringBufs := NewRingBuffers(cmdStrings...)

	type args struct {
		cmd string
		n   int
	}
	tests := []struct {
		name    string
		r       RingBuffers
		args    args
		want    []byte
		wantErr bool
	}{
		{name: "TestRingBuffers_Read #1", r: ringBufs, args: args{cmd: "test1", n: 3}, want: []byte{0x01, 0x02, 0x03}, wantErr: false},
		{name: "TestRingBuffers_Read #2", r: ringBufs, args: args{cmd: "test2", n: 2}, want: []byte{0x04, 0x05}, wantErr: false},
	}

	ringBufs["test1"].EnqueueN([]byte{0x01, 0x02, 0x03}, 3)
	ringBufs["test2"].EnqueueN([]byte{0x04, 0x05}, 2)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.Read(tt.args.cmd, tt.args.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("RingBuffers.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RingBuffers.Read() = %v, want %v", got, tt.want)
			}
		})
	}
}
