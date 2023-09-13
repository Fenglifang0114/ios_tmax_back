package svc

import "testing"

func Test_getPercentage(t *testing.T) {
	type args struct {
		data string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "Test_getPercentage", args: args{data: `b\b \b[ 10%]\b \b\b \b\b \b\b \b\b \b\b \b[ 11%]\b \b\b \b\b \b\b \b\b \b\b \b[ 11%]\b \b\b \b\b \b\b \b\b \b\b \b[ 11%]\b \b\b \b\b \b\b \b\b \b\b \b[ 12%]\b \b\b \b\b \b\b \b\b \b\b \b[ 12%]\b \b\b \b\b \b\b \b\b \b\b \b[ 13%]\b \b\b \b\b \b\b \b\b \b\b \b[ 13%]\b \b\b \b\b \b\b \b\b \b\b \b[ 13%]\b \b\b \b\b \b\b \b\b \b\b \b[ 14%]\b \b\b \b\b \b\b \`}, want: "14%"},
		{name: "Test_getPercentage", args: args{data: `b\b \b[ 10%]\b \b\b \b\b \b\b \b\b \b\b \b[ 11%]\b \b\b \b\b \b\b \b\b \b\b \b[ 11%]\b \b\b \b\b \b\b \b\b \b\b \b[ 11%]\b \b\b \b\b \b\b \b\b \b\b \b[ 12%]\b \b\b \b\b \b\b \b\b \b\b \b[ 12%]\b \b\b \b\b \b\b \b\b \b\b \b[ 13%]\b \b\b \b\b \b\b \b\b \b\b \b[ 13%]\b \b\b \b\b \b\b \b\b \b\b \b[ 13%]\b \b\b \b\b \b\b \b\b \b\b \b[100%]\b \b\b \b\b \b\b \`}, want: "100%"},
		{name: "Test_getPercentage", args: args{data: `b\b \b[ 10%]\b \b\b \b\b \b\b \b\b \b\b \b[ 11%]\b \b\b \b\b \b\b \b\b \b\b \b[ 11%]\b \b\b \b\b \b\b \b\b \b\b \b[ 11%]\b \b\b \b\b \b\b \b\b \b\b \b[ 12%]\b \b\b \b\b \b\b \b\b \b\b \b[ 12%]\b \b\b \b\b \b\b \b\b \b\b \b[ 13%]\b \b\b \b\b \b\b \b\b \b\b \b[ 13%]\b \b\b \b\b \b\b \b\b \b\b \b[ 13%]\b \b\b \b\b \b\b \b\b \b\b \b[  1%]\b \b\b \b\b \b\b \`}, want: "1%"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPercentage(tt.args.data); got != tt.want {
				t.Errorf("getPercentage() = %v, want %v", got, tt.want)
			}
		})
	}
}
