package authorization

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	mock_authorization "github.com/pjoc-team/pay-gateway/pkg/authorization/mock"
)

func Test_parseAuthData(t *testing.T) {
	type args struct {
		authData string
	}
	tests := []struct {
		name    string
		args    args
		want    *AuthInfo
		wantErr bool
	}{
		{
			name: "expect",
			args: args{
				authData: `merchant_id="1900009191",nonce="593BEC0C930BF1AFEB40B4A08C8FB242",signature="uOVRnA4qG/MNnYzdQxJanN+zU+lTgIcnU9BxGw5dKjK+VdEUz2FeIoC+D5sB/LN+nGzX3hfZg6r5wT1pl2ZobmIc6p0ldN7J6yDgUzbX8Uk3sD4a4eZVPTBvqNDoUqcYMlZ9uuDdCvNv4TM3c1WzsXUrExwVkI1XO5jCNbgDJ25nkT/c1gIFvqoogl7MdSFGc4W4xZsqCItnqbypR3RuGIlR9h9vlRsy7zJR9PBI83X8alLDIfR1ukt1P7tMnmogZ0cuDY8cZsd8ZlCgLadmvej58SLsIkVxFJ8XyUgx9FmutKSYTmYtWBZ0+tNvfGmbXU7cob8H/4nLBiCwIUFluw==",timestamp="1554208460",serial_no="1DDE55AD98ED71D6EDD4A4A16996DE7B47773A8C"`,
			},
			want: &AuthInfo{
				Timestamp:  "1554208460",
				Nonce:      "593BEC0C930BF1AFEB40B4A08C8FB242",
				Signature:  "uOVRnA4qG/MNnYzdQxJanN+zU+lTgIcnU9BxGw5dKjK+VdEUz2FeIoC+D5sB/LN+nGzX3hfZg6r5wT1pl2ZobmIc6p0ldN7J6yDgUzbX8Uk3sD4a4eZVPTBvqNDoUqcYMlZ9uuDdCvNv4TM3c1WzsXUrExwVkI1XO5jCNbgDJ25nkT/c1gIFvqoogl7MdSFGc4W4xZsqCItnqbypR3RuGIlR9h9vlRsy7zJR9PBI83X8alLDIfR1ukt1P7tMnmogZ0cuDY8cZsd8ZlCgLadmvej58SLsIkVxFJ8XyUgx9FmutKSYTmYtWBZ0+tNvfGmbXU7cob8H/4nLBiCwIUFluw==",
				SerialNo:   "1DDE55AD98ED71D6EDD4A4A16996DE7B47773A8C",
				MerchantID: "1900009191",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := parseAuthInfo(tt.args.authData)
				if (err != nil) != tt.wantErr {
					t.Errorf("parseAuthInfo() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("parseAuthInfo() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_authInterceptor_verifyAuthorization(t *testing.T) {
	httpBody := `{"order_id": "123"}`
	merchantID := "m1"
	serialNO := "s1"
	certificate, err := GenerateCertificate()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(certificate.SerialNumber)

	ctrl := gomock.NewController(t)
	manager := mock_authorization.NewMockCertificateManager(ctrl)
	manager.EXPECT().GetMerchantCertificate(context.TODO(), merchantID, serialNO).Return(certificate, nil)

	type fields struct {
		certificateManager CertificateManager
	}
	type args struct {
		ctx             context.Context
		authHeader      string
		httpMethod      string
		httpPath        string
		httpRequestBody []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "",
			fields: fields{
				certificateManager: nil,
			},
			args: args{
				ctx:             nil,
				authHeader:      "",
				httpMethod:      "",
				httpPath:        "",
				httpRequestBody: []byte(httpBody),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				a := &authInterceptor{
					certificateManager: tt.fields.certificateManager,
				}
				if err := a.verifyAuthorization(
					tt.args.ctx, tt.args.authHeader, tt.args.httpMethod, tt.args.httpPath, tt.args.httpRequestBody,
				); (err != nil) != tt.wantErr {
					t.Errorf("verifyAuthorization() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}
