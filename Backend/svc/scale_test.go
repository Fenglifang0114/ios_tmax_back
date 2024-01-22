package svc

import (
	"reflect"
	"testing"
	"tmaxsrv/comm"
)

func TestNewScale(t *testing.T) {
	mgr := NewScaleMgr()
	mediaConf := `{"Mode":{"BaudRate":9600,"DataBits":8,"Parity":0,"StopBits":0,"InitialStatusBits":null},"PortName":"COM6"}`
	scaleConn := &ScaleConnMedia{ScaleModel: "ATP", ScaleSn: "123456", TMedia: MEDIA_COM, MediaConf: MediaConf{Type: MEDIA_COM, MediaInfoJson: mediaConf}}
	type args struct {
		scaleMgr *ScaleMgr
		conn     *ScaleConnMedia
		model    string
		sn       string
	}
	tests := []struct {
		name    string
		args    args
		want    *Scale
		wantErr bool
	}{
		{name: "new scale test", args: args{scaleMgr: mgr, conn: scaleConn, model: ""}, want: nil, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewScale(tt.args.scaleMgr, scaleConn, comm.SCALE_T2200, tt.args.model, tt.args.sn, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewScale() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Errorf("NewScale() return nil")
			}
		})
	}
}

// func Test_readScale(t *testing.T) {
// 	scaleMgr := NewScaleMgr()
// 	mediaConf := `{"Mode":{"BaudRate":9600,"DataBits":8,"Parity":0,"StopBits":0,"InitialStatusBits":null},"PortName":"COM6"}`
// 	scaleConn := &ScaleConnMedia{ScaleModel: "ATP", ScaleSn: "123456", TMedia: MEDIA_COM, MediaConf: MediaConf{Type: MEDIA_COM, MediaInfoJson: mediaConf}}
// 	myscale, _ := NewScale(scaleMgr, scaleConn, "ATP", "123456", true)

// 	type args struct {
// 		c *Scale
// 	}

// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    []byte
// 		wantErr bool
// 	}{
// 		{name: "read scale", args: args{c: myscale}, want: []byte("1234567890\r\n"), wantErr: false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			err := readScale(tt.args.c)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("readScale() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			// if !reflect.DeepEqual(got, tt.want) {
// 			// 	t.Errorf("readScale() = %v, want %v", got, tt.want)
// 			// }
// 		})
// 	}
// }

// func Test_packMsg(t *testing.T) {
// 	type args struct {
// 		data          []byte
// 		isOldC51Scale bool
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    *ScaleMsg
// 		wantErr bool
// 	}{
// 		{name: "packMsg", args: args{data: []byte("ST,NT,5.123kg"), isOldC51Scale: true},
// 			want: &ScaleMsg{MsgType: WEIGHT_DATA, MsgBody: WeightMsg{IsStable: true, IsNet: true, WeightVal: "5.123", WeightUnit: "kg"}}, wantErr: false},
// 		{name: "packMsg", args: args{data: []byte("UN,GS,50.4g"), isOldC51Scale: true},
// 			want: &ScaleMsg{MsgType: WEIGHT_DATA, MsgBody: WeightMsg{IsStable: false, IsNet: false, WeightVal: "50.4", WeightUnit: "g"}}, wantErr: false},
// 		{name: "packMsg", args: args{data: []byte("UN,GS,5h0.4g"), isOldC51Scale: true},
// 			want: &ScaleMsg{MsgType: WEIGHT_DATA, MsgBody: WeightMsg{IsStable: false, IsNet: false, WeightVal: "", WeightUnit: ""}}, wantErr: true},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := packMsg(tt.args.data, tt.args.isOldC51Scale)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("packMsg() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("packMsg() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func Test_writeScale(t *testing.T) {
	// prepare scale instance
	scaleMgr := NewScaleMgr()
	mediaConf := `{"Mode":{"BaudRate":9600,"DataBits":8,"Parity":0,"StopBits":0,"InitialStatusBits":null},"PortName":"COM6"}`
	scaleConn := &ScaleConnMedia{ScaleModel: "ATP", ScaleSn: "123456", TMedia: MEDIA_COM, MediaConf: MediaConf{Type: MEDIA_COM, MediaInfoJson: mediaConf}}
	myscale, _ := NewScale(scaleMgr, scaleConn, comm.SCALE_T2200, "ATP", "123456", true)

	type args struct {
		c    *Scale
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{name: "scale", args: args{c: myscale, data: []byte("foo\r\n")}, want: 5, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := writeScale(tt.args.c, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("writeScale() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if got != tt.want {
			// 	t.Errorf("writeScale() = %v, want %v", got, tt.want)
			// }
		})
	}
}

func Test_remove(t *testing.T) {
	ch1 := make(chan *ScaleRespMsg)
	ch2 := make(chan *ScaleRespMsg)
	ch3 := make(chan *ScaleRespMsg)
	type args struct {
		s []chan *ScaleRespMsg
		m chan *ScaleRespMsg
	}
	tests := []struct {
		name string
		args args
		want []chan *ScaleRespMsg
	}{
		{name: "remove test", args: args{s: []chan *ScaleRespMsg{ch1, ch2, ch3}, m: ch2}, want: []chan *ScaleRespMsg{ch1, ch3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := remove(tt.args.s, tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("remove() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScale_RegisterNotif(t *testing.T) {
	// prepare scale instance
	scaleMgr := NewScaleMgr()
	mediaConf := `{"Mode":{"BaudRate":9600,"DataBits":8,"Parity":0,"StopBits":0,"InitialStatusBits":null},"PortName":"COM6"}`
	scaleConn := &ScaleConnMedia{ScaleModel: "ATP", ScaleSn: "123456", TMedia: MEDIA_COM, MediaConf: MediaConf{Type: MEDIA_COM, MediaInfoJson: mediaConf}}
	myscale, _ := NewScale(scaleMgr, scaleConn, comm.SCALE_T2200, "ATP", "123456", true)

	type args struct {
		msgType comm.RespMsgType
		inCh    chan *ScaleRespMsg
	}
	tests := []struct {
		name string
		c    *Scale
		args args
	}{
		{name: "RegisterNotif test", c: myscale, args: args{msgType: comm.WEIGHT_DATA, inCh: make(chan *ScaleRespMsg)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.RegisterNotif(tt.args.msgType, tt.args.inCh)
		})
	}
}

func TestScale_UnRegisterNotif(t *testing.T) {
	// prepare scale instance
	scaleMgr := NewScaleMgr()
	_ = NewSrvMgr(scaleMgr, make(chan bool))
	mediaConf := `{"Mode":{"BaudRate":9600,"DataBits":8,"Parity":0,"StopBits":0,"InitialStatusBits":null},"PortName":"COM6"}`
	scaleConn := &ScaleConnMedia{ScaleModel: "ATP", ScaleSn: "123456", TMedia: MEDIA_COM, MediaConf: MediaConf{Type: MEDIA_COM, MediaInfoJson: mediaConf}}
	myscale, _ := NewScale(scaleMgr, scaleConn, comm.SCALE_T2200, "ATP", "123456", true)

	type args struct {
		msgType comm.RespMsgType
		inCh    chan *ScaleRespMsg
	}
	tests := []struct {
		name string
		c    *Scale
		args args
	}{
		{name: "RegisterNotif test", c: myscale, args: args{msgType: comm.WEIGHT_DATA, inCh: make(chan *ScaleRespMsg)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.RegisterNotif(tt.args.msgType, tt.args.inCh)
			tt.c.UnRegisterNotif(tt.args.msgType, tt.args.inCh)
			if len(tt.c.respChansMap[tt.args.msgType]) != 0 {
				t.Errorf("UnRegisterNotif() fail!")
			}
		})
	}
}

func Test_retreiveWeight(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    WeightMsg
		wantErr bool
	}{
		{name: "retreive weight #1", args: args{data: []byte("ST,NT, 5.123kg")}, want: WeightMsg{IsStable: true, IsNet: true, WeightVal: "5.123", WeightUnit: "kg"}, wantErr: false},
		{name: "retreive weight #2", args: args{data: []byte("ST,NT,5.123kg")}, want: WeightMsg{IsStable: true, IsNet: true, WeightVal: "5.123", WeightUnit: "kg"}, wantErr: false},
		{name: "retreive weight #2", args: args{data: []byte("ST,NT,-5.123kg")}, want: WeightMsg{IsStable: true, IsNet: true, WeightVal: "-5.123", WeightUnit: "kg"}, wantErr: false},
		{name: "retreive weight #2", args: args{data: []byte("US,GS,-5.123kg")}, want: WeightMsg{IsStable: false, IsNet: false, WeightVal: "-5.123", WeightUnit: "kg"}, wantErr: false},
		{name: "retreive weight #2", args: args{data: []byte("US,GS, 5.123g")}, want: WeightMsg{IsStable: false, IsNet: false, WeightVal: "5.123", WeightUnit: "g"}, wantErr: false},
		{name: "retreive weight #3", args: args{data: []byte("-- UL --")}, want: WeightMsg{IsStable: false, IsNet: false, WeightVal: "--UL--", WeightUnit: ""}, wantErr: false},
		{name: "retreive weight #4", args: args{data: []byte("1234567890")}, want: WeightMsg{IsStable: false, IsNet: false, WeightVal: "1234567890", WeightUnit: ""}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := retreiveWeightC51(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("retreiveWeight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("retreiveWeight() = %v, want %v", got, tt.want)
			}
		})
	}
}
