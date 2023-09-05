package svc

import (
	"reflect"
	"testing"
)

func Test_handlePortState(t *testing.T) {
	jsonStr1, _ := json.MarshalToString(ComInfo{DevPath: "COM2"})
	jsonStr2, _ := json.MarshalToString(ComInfo{DevPath: "COM5"})
	type args struct {
		inPorts *[]string
		conns   *[]*ScaleConnMedia
	}
	tests := []struct {
		name              string
		args              args
		wantPortsNotInUse []string
	}{
		{
			name: "handlePortState #1", args: args{
				inPorts: &[]string{"COM1", "COM2", "COM3"},
				conns:   &[]*ScaleConnMedia{{IsOnline: true, TMedia: MEDIA_COM, MediaConf: MediaConf{Type: MEDIA_COM, MediaInfoJson: jsonStr1}}},
			},
			wantPortsNotInUse: []string{"COM1", "COM3"},
		},
		{
			name: "handlePortState #2", args: args{
				inPorts: &[]string{"COM1", "COM2", "COM3"},
				conns:   &[]*ScaleConnMedia{{IsOnline: true, TMedia: MEDIA_COM, MediaConf: MediaConf{Type: MEDIA_COM, MediaInfoJson: jsonStr2}}},
			},
			wantPortsNotInUse: []string{"COM1", "COM2", "COM3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotPortsNotInUse := handlePortState(tt.args.inPorts, tt.args.conns); !reflect.DeepEqual(gotPortsNotInUse, tt.wantPortsNotInUse) {
				t.Errorf("handlePortState() = %v, want %v", gotPortsNotInUse, tt.wantPortsNotInUse)
			}
		})
	}
}
