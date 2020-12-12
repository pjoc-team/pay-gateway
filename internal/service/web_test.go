package service

import (
	"fmt"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"testing"
)

func Test_decode(t *testing.T) {
	type args struct {
		rawData []byte
		v       interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "t1",
			args: args{
				rawData: []byte("hello"),
				v:       &httpbody.HttpBody{},
			},
			wantErr: false,
		},
		{
			name: "t2",
			args: args{
				rawData: []byte("hello"),
				v:       &([]byte{}),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if err := decode(tt.args.rawData, tt.args.v); (err != nil) != tt.wantErr {
					t.Errorf("decode() error = %v, wantErr %v", err, tt.wantErr)
				}
				fmt.Println(tt.args.v)
			},
		)
	}
}
