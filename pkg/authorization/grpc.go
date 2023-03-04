package authorization

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/blademainer/commons/pkg/field"
	"github.com/pjoc-team/pay-gateway/pkg/grpc/interceptors/http"
	"github.com/pjoc-team/pay-gateway/pkg/grpc/metadata"
	"github.com/pjoc-team/tracing/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	// "google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const authMethodHeader = "PJOCPAY-SHA256-RSA2048"
const timestampHeader = "timestamp"
const nonceHeader = "nonce"
const signHeader = "signature"
const SerialNoHeader = "serial_no"

type authInterceptor struct {
	apiKey          string // 微信支付分配给商户的apiKey
	apiSecret       string // 微信支付分配给商户的apiSecret
	mchID           string // 微信支付分配给商户的mchID
	serialNumber    string // 商户证书序列号
	privateKey      []byte // 商户私钥
	certificateList []*x509.Certificate
}

func newAuthInterceptor(apiKey, apiSecret, mchID, serialNumber string, privateKey []byte) *authInterceptor {
	return &authInterceptor{
		apiKey:       apiKey,
		apiSecret:    apiSecret,
		mchID:        mchID,
		serialNumber: serialNumber,
		privateKey:   privateKey,
	}
}

func (a *authInterceptor) UnaryServerInterceptor(
	ctx context.Context, req string, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (interface{}, error) {
	log := logger.ContextLog(ctx)
	md := metadata.FromIncomingContext(ctx)
	authHeader := md.GetAuthorization()
	if authHeader == "" {
		return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	authFields := strings.Split(authHeader, " ")
	if len(authFields) != 2 || authFields[0] != authMethodHeader {
		return nil, status.Errorf(codes.Unauthenticated, "invalid authorization token")
	}

	// 解析Authorization头部的值
	authData, err := url.QueryUnescape(authFields[1])
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to decode authorization data")
	}

	authInfo, err := parseAuthInfo(authData)
	if err != nil {
		log.Errorf("failed to parse auth info: %v, error: %v", authData, err.Error())
		return nil, err
	}

	timestamp, err := strconv.ParseInt(authInfo.Timestamp, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid timestamp")
	}

	// 判断请求是否过期
	if time.Now().Unix()-timestamp > 300 {
		return nil, status.Errorf(codes.Unauthenticated, "authorization token has expired")
	}

	// 从请求中获取body
	body, ok := ctx.Value(http.ContextHttpRequestBody).([]byte)
	if !ok {
		return nil, status.Errorf(codes.Internal, "failed to get request body")
	}
	// 从请求中获取method
	method, ok := ctx.Value(http.ContextHttpRequestMethod).(string)
	if !ok {
		return nil, status.Errorf(codes.Internal, "failed to get request method")
	}

	// 构造待签名字符串
	// 签名串一共有五行，每一行为一个参数。行尾以 \n（换行符，ASCII编码值为0x0A）结束，包括最后一行。如果参数本身以\n结束，也需要附加一个\n。
	//
	//
	//					  HTTP请求方法\n
	//					  URL\n
	//					  请求时间戳\n
	//					  请求随机串\n
	//					  请求报文主体\n
	signStr := fmt.Sprintf("%s\n%s\n%d\n%s\n%s\n", method, info.FullMethod, timestamp, authInfo.Nonce, body)

	// 计算签名
	signatureBytes, err := base64.StdEncoding.DecodeString(authInfo.Signature)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to decode signature: %v", err)
	}
	var cert *x509.Certificate
	for _, c := range a.certificateList {
		if c.SerialNumber.String() == a.serialNumber {
			cert = c
			break
		}
	}
	if cert == nil {
		return nil, status.Errorf(codes.Unauthenticated, "certificate not found")
	}
	hashed := sha256.Sum256([]byte(signStr))
	err = rsa.VerifyPKCS1v15(cert.PublicKey.(*rsa.PublicKey), crypto.SHA256, hashed[:], signatureBytes)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid signature: %v", err)
	}

	return handler(ctx, req)
}

var p = &field.Parser{
	Tag:                 "json",
	Escape:              true,
	GroupDelimiter:      ',',
	PairDelimiter:       '=',
	Sort:                false,
	IgnoreNilValueField: true,
}

func parseAuthInfo(authData string) (*AuthInfo, error) {
	authParams := make(map[string][]string)
	for _, kv := range strings.Split(authData, ",") {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			return nil, status.Errorf(codes.Unauthenticated, "invalid authorization data format")
		}
		authParams[parts[0]] = []string{strings.ReplaceAll(parts[1], "\"", "")}
	}
	auth := &AuthInfo{}
	err := p.Unmarshal(auth, authParams)
	if err != nil {
		return nil, err
	}
	return auth, nil
}
