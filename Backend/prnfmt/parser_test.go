package prnfmt

import (
	"bytes"
	"testing"
)

func TestParserFmtToBuf(t *testing.T) {
	var mybuf bytes.Buffer
	mybuf.WriteString("Hello, World!")
	type args struct {
		utf8Buff string
	}
	tests := []struct {
		name string
		args args
		want *bytes.Buffer
	}{
		{name: "TestParserFmtToBuf", args: args{utf8Buff: `P,396,360
TB,29,38,130,30,0,1,1,0,0,DATA,Date,2023-03-16,1,10,2
TB,165,38,130,30,0,1,1,0,0,DATA,Time,10:38:16,1,10,3
TB,100,116,121,30,0,1,1,0,0,DATA,percent,percent,3,7,13
TB,29,116,109,30,0,1,1,0,0,TEXT,PERCENT:,14
TB,227,116,130,30,0,1,1,0,0,TEXT,%,15
ROTATE,0`}, want: &mybuf},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParserFmtToBuf(tt.args.utf8Buff, "EPM205"); len(got.Bytes()) == 2048 {
				t.Errorf("ParserFmtToBuf() = %s,", got.Bytes())
			}
		})
	}
}
