package svc

import "testing"

func TestDbProductRec_UpdateProductRec(t *testing.T) {
	db, _ := NewDbProductRec("testproductrec.db")
	type args struct {
		rec ProductRec
	}
	tests := []struct {
		name    string
		d       *DbProductRec
		args    args
		wantErr bool
	}{
		{name: "UpdateScaleRec", d: db, args: args{rec: ProductRec{RecId: 22, Id: "2222", WithPretare: true, Product: "2222", Pretare: "222", Remarks: "2222"}}, wantErr: false},
		{name: "UpdateScaleRec", d: db, args: args{rec: ProductRec{RecId: 1, Id: "1111", WithPretare: false, Product: "1111", Pretare: "1111", Remarks: "1111"}}, wantErr: false},
		{name: "UpdateScaleRec1", d: db, args: args{rec: ProductRec{RecId: 888, Id: "22", WithPretare: true, Product: "92229", Pretare: "222", Remarks: "228"}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.UpdateProductRec(tt.args.rec); (err != nil) != tt.wantErr {
				t.Errorf("DbProductRec.UpdateProductRec() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
