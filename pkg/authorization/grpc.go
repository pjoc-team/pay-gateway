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

	"github.com/pjoc-team/pay-gateway/pkg/grpc/metadata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	// "google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type authInterceptor struct {
	apiKey       string // 微信支付分配给商户的apiKey
	apiSecret    string // 微信支付分配给商户的apiSecret
	mchID        string // 微信支付分配给商户的mchID
	serialNumber string // 商户证书序列号
	privateKey   []byte // 商户私钥
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
	md := metadata.FromIncomingContext(ctx)
	authHeader := md.GetAuthorization()
	if authHeader == "" {
		return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	authFields := strings.Split(authHeader, " ")
	if len(authFields) != 2 || authFields[0] != "WECHATPAY2-SHA256-RSA2048" {
		return nil, status.Errorf(codes.Unauthenticated, "invalid authorization token")
	}

	// 解析Authorization头部的值
	authData, err := url.QueryUnescape(authFields[1])
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to decode authorization data")
	}

	authParams := make(map[string]string)
	for _, kv := range strings.Split(authData, "&") {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			return nil, status.Errorf(codes.Unauthenticated, "invalid authorization data format")
		}
		authParams[parts[0]] = parts[1]
	}

	timestamp, err := strconv.ParseInt(authParams["timestamp"], 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid timestamp")
	}

	// 判断请求是否过期
	if time.Now().Unix()-timestamp > 300 {
		return nil, status.Errorf(codes.Unauthenticated, "authorization token has expired")
	}

	nonce := authParams["nonce"]
	signature := authParams["signature"]

	// 从请求中获取body
	body, err := grpc_getRequestBody(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get request body: %v", err)
	}

	// 构造待签名字符串
	signStr := fmt.Sprintf("%d\n%s\n%s\n%s\n%s\n", timestamp, nonce, info.FullMethod, body, a.mchID)

	// 计算签名
	signatureBytes, err := base64.StdEncoding.DecodeString(signature)
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
