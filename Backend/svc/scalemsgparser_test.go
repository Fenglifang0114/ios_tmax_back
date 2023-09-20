package svc

import (
	"fmt"
	"reflect"
	"testing"
	"tmaxsrv/comm"
)

func Test_parseNetworkInfo(t *testing.T) {
	type args struct {
		response []byte
	}
	tests := []struct {
		name string
		args args
		want []CWLAPResponse
	}{
		{
			name: "Test_parseNetworkInfo",
			args: args{[]byte("+CWLAP:(3,\"Test1\",-92,\"bc:e2:65:c1:22:b0\",1,18,2)\r\n+CWLAP:(3,\"Test2\",-92,\"bc:e2:65:c1:22:bb\",1,18,2)\r\n\r\nOK\r\n")},
			want: []CWLAPResponse{{3, "Test1", -92, "bc:e2:65:c1:22:b0", 1, 18, 2}, {3, "Test2", -92, "bc:e2:65:c1:22:bb", 1, 18, 2}},
		},
		{
			name: "Test_parseNetworkInfo",
			args: args{[]byte("+CWLAP:(3,\"Test2\",-88,\"bc:e2:65:c1:22:b2\",2,19,0)\r\n")},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseCWLAPResponse(tt.args.response); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseNetworkInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertResponsesToInfos(t *testing.T) {
	type args struct {
		responses []CWLAPResponse
	}
	tests := []struct {
		name string
		args args
		want []APInfo
	}{
		{
			name: "Test_convertResponsesToInfos",
			args: args{[]CWLAPResponse{{2, "test1", -65, "11:22:33:44:55:66", 2, 66, 1}, {2, "test2", -66, "11:22:33:44:55:68", 3, 67, 2}}},
			want: []APInfo{{0, "test1", 3, "11:22:33:44:55:66", "NONE"}, {1, "test2", 3, "11:22:33:44:55:68", "WEP"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertResponsesToInfos(tt.args.responses); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertResponsesToInfos() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleGetApListResp(t *testing.T) {
	type args struct {
		scaleId int64
		data    []byte
	}
	tests := []struct {
		name  string
		args  args
		want  ScaleRespMsg
		want1 int
	}{
		{name: "Test_handleGetApListResp", args: args{1, []byte("+CWLAP:(3,\"Test1\",-92,\"bc:e2:65:c1:22:b0\",1,18,2)\r\n+CWLAP:(3,\"Test2\",-92,\"bc:e2:65:c1:22:bb\",1,18,2)\r\n\r\nOK\r\n")}, want: ScaleRespMsg{comm.GET_AP_LIST_RESP, "[{\"seqno\":0,\"ssid\":\"Test1\",\"rssi\":1,\"mac\":\"bc:e2:65:c1:22:b0\"},{\"seqno\":1,\"ssid\":\"Test2\",\"rssi\":1,\"mac\":\"bc:e2:65:c1:22:bb\"}]", 1}, want1: 108},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := handleGetApListResp(tt.args.scaleId, tt.args.data)
			fmt.Printf("%v", got.MsgBody)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleGetApListResp() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("handleGetApListResp() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_getRssiLevel(t *testing.T) {
	type args struct {
		rssi int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "Test_getRssiLevel #1", args: args{rssi: -80}, want: 2},
		{name: "Test_getRssiLevel #2", args: args{rssi: -100}, want: 1},
		{name: "Test_getRssiLevel #3", args: args{rssi: -55}, want: 4},
		{name: "Test_getRssiLevel #4", args: args{rssi: -70}, want: 3},
		{name: "Test_getRssiLevel #5", args: args{rssi: -40}, want: 4},
		{name: "Test_getRssiLevel #5", args: args{rssi: -200}, want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRssiLevel(tt.args.rssi); got != tt.want {
				t.Errorf("getRssiLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractWifiAPInfo(t *testing.T) {
	type args struct {
		response string
	}
	tests := []struct {
		name    string
		args    args
		want    WifiAPInfo
		wantErr bool
	}{
		{name: "Test_extractWifiAPInfo #1", args: args{response: `AT+CWJAP_DEF?
  
		+CWJAP_DEF:"T-Scale","0c:4b:54:61:fd:db",1,-67
		
		OK`}, want: WifiAPInfo{"T-Scale", "0c:4b:54:61:fd:db", "1", 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractWifiAPInfo(tt.args.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractWifiAPInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractWifiAPInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_retrieveWeight(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    WeightMsg
		wantErr bool
	}{
		{name: "Test_retrieveWeight #1", args: args{data: []byte("ST,GS,-0.123kg\r\n")}, want: WeightMsg{true, false, "-0.123", "kg"}, wantErr: false},
		{name: "Test_retrieveWeight #2", args: args{data: []byte("UT,NT,-0.123kg\r\n")}, want: WeightMsg{false, true, "-0.123", "kg"}, wantErr: false},
		{name: "Test_retrieveWeight #3", args: args{data: []byte("ST,GS, 90pcs\r\n")}, want: WeightMsg{true, false, "90", "pcs"}, wantErr: false},
		{name: "Test_retrieveWeight #3", args: args{data: []byte("ST,GS,90    pc\r\n")}, want: WeightMsg{true, false, "90", "pc"}, wantErr: false},
		{name: "Test_retrieveWeight #5", args: args{data: []byte("ST,GS, 89%\r\n")}, want: WeightMsg{true, false, "89", "%"}, wantErr: false},
		{name: "Test_retrieveWeight #6", args: args{data: []byte("ST,GS,100pcs\r\n")}, want: WeightMsg{true, false, "100", "pcs"}, wantErr: false},
		{name: "Test_retrieveWeight #7", args: args{data: []byte("ST,GS,100%\r\n")}, want: WeightMsg{true, false, "100", "%"}, wantErr: false},
		{name: "Test_retrieveWeight #8", args: args{data: []byte("ST,GS,-0.123kg\r\n")}, want: WeightMsg{true, false, "-0.123", "kg"}, wantErr: false},
		{name: "Test_retrieveWeight #9", args: args{data: []byte("--OL--        \r\n")}, want: WeightMsg{false, false, "--OL--", ""}, wantErr: false},
		{name: "Test_retrieveWeight #10", args: args{data: []byte("--UL--        \r\n")}, want: WeightMsg{false, false, "--UL--", ""}, wantErr: false},
		{name: "Test_retrieveWeight #11", args: args{data: []byte("ST,GS-0.123kg\r\n")}, want: WeightMsg{false, false, "", ""}, wantErr: true},
		{name: "Test_retrieveWeight #12", args: args{data: []byte("UT,NT-0.123kg\r\n")}, want: WeightMsg{false, false, "", ""}, wantErr: true},
		{name: "Test_retrieveWeight #13", args: args{data: []byte("UT,NT,253:0.12lb\r\n")}, want: WeightMsg{false, true, "253:0.12", "lb"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := retrieveWeight(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("retrieveWeight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("retrieveWeight() = %v, want %v", got, tt.want)
			}
		})
	}
}
