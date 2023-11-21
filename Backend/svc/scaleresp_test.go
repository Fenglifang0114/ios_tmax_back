package svc

import (
	"reflect"
	"testing"
)

func Test_retreiveWeightC51(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name     string
		args     args
		wantPack WeightMsg
		wantErr  bool
	}{
		{name: "Test_retreiveWeightC51 #1", args: args{[]byte("ZE,ST,NET, 1.000kg")}, wantPack: WeightMsg{true, true, false, "1.000", "kg"}, wantErr: false},
		{name: "Test_retreiveWeightC51 #2", args: args{[]byte("ZE,ST,NT, 1.000kg")}, wantPack: WeightMsg{true, true, true, "1.000", "kg"}, wantErr: false},
		{name: "Test_retreiveWeightC51 #3", args: args{[]byte("ZE,UT,GS, 1.000kg")}, wantPack: WeightMsg{true, false, false, "1.000", "kg"}, wantErr: false},
		{name: "Test_retreiveWeightC51 #4", args: args{[]byte("ZE,ST,NET,0.123g")}, wantPack: WeightMsg{true, true, false, "0.123", "g"}, wantErr: false},
		{name: "Test_retreiveWeightC51 #5", args: args{[]byte("ZE,ST,NT, 100pcs")}, wantPack: WeightMsg{true, true, true, "100", "pcs"}, wantErr: false},
		{name: "Test_retreiveWeightC51 #6", args: args{[]byte(`ZE,ST,NET, 90.23%`)}, wantPack: WeightMsg{true, true, false, "90.23", "%"}, wantErr: false},
		{name: "Test_retreiveWeightC51 #7", args: args{[]byte(`-- OL --`)}, wantPack: WeightMsg{true, false, false, "-- OL --", ""}, wantErr: false},
		{name: "Test_retreiveWeightC51 #8", args: args{[]byte(`-- UL --`)}, wantPack: WeightMsg{true, false, false, "-- UL --", ""}, wantErr: false},
		{name: "Test_retreiveWeightC51 #9", args: args{[]byte(`-- UL --\r\n`)}, wantPack: WeightMsg{true, false, false, "-- UL --", ""}, wantErr: false},
		{name: "Test_retreiveWeightC51 #9", args: args{[]byte(`-- UL --\r\n-- UL --\r\n`)}, wantPack: WeightMsg{true, false, false, "-- UL --", ""}, wantErr: false},
		{name: "Test_retreiveWeightC51 #10", args: args{[]byte(`932854ldskjfglksdjfgkljsklfgjkl`)}, wantPack: WeightMsg{}, wantErr: true},
		{name: "Test_retreiveWeightC51 #11", args: args{[]byte("ZE,ST,NET, 1.000lb")}, wantPack: WeightMsg{true, true, false, "1.000", "lb"}, wantErr: false},
		{name: "Test_retreiveWeightC51 #12", args: args{[]byte("ZE,ST,NT, 1.000lboz")}, wantPack: WeightMsg{true, true, true, "1.000", "lboz"}, wantErr: false},
		{name: "Test_retreiveWeightC51 #13", args: args{[]byte("ZE,UT,GS, 1.000tj")}, wantPack: WeightMsg{true, false, false, "1.000", "tj"}, wantErr: false},
		{name: "Test_retreiveWeightC51 #14", args: args{[]byte("ZE,ST,NET,0.123hj")}, wantPack: WeightMsg{true, true, false, "0.123", "hj"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPack, err := retreiveWeightC51(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("retreiveWeightC51() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPack, tt.wantPack) {
				t.Errorf("retreiveWeightC51() = %v, want %v", gotPack, tt.wantPack)
			}
		})
	}
}
