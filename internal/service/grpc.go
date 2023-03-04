package service

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pjoc-team/pay-gateway/pkg/grpc/metadata"
	"google.golang.org/grpc"
)

var (
	// GrpcServices 需要注册grpc的服务
	GrpcServices = make(map[string]GrpcInfo)
)

// RegisterGrpcFunc 注册grpc函数
type RegisterGrpcFunc func(ctx context.Context, server *grpc.Server) error

// RegisterGatewayFunc 注册gatway函数
type RegisterGatewayFunc func(ctx context.Context, mux *runtime.ServeMux) error

// RegisterStreamFunc 注册stream
type RegisterStreamFunc func(
	ctx context.Context, mux *runtime.ServeMux,
	conn *grpc.ClientConn,
) error

// GrpcInfo 注册grpc服务的函数。请在该函数被调用时注册自己的grpc服务。
// 在服务启动时，会调用所有该类型函数。
type GrpcInfo struct {
	RegisterGrpcFunc    RegisterGrpcFunc
	RegisterGatewayFunc RegisterGatewayFunc
	RegisterStreamFunc  RegisterStreamFunc

	Name string
}

// RegisterGrpc 注册grpc服务。这里会同时将grpc服务注册到grpc和gateway
func RegisterGrpc(
	name string, grpcFunc RegisterGrpcFunc, gatewayFunc RegisterGatewayFunc,
	streamFunc RegisterStreamFunc,
) {
	GrpcServices[name] = GrpcInfo{
		Name:                name,
		RegisterGrpcFunc:    grpcFunc,
		RegisterGatewayFunc: gatewayFunc,
		RegisterStreamFunc:  streamFunc,
	}
}

func newGrpcMux() *runtime.ServeMux {
	// init grpc gateway
	// marshaler := &runtime.JSONPb{
	// 	EnumsAsInts:  false, // 枚举类使用string返回
	// 	OrigName:     true,  // 使用json tag里面的字段
	// 	EmitDefaults: true,  // json返回零值
	// }
	// marshalOpts := runtime.WithMarshalerOption(runtime.MIMEWildcard, marshaler)

	mux := runtime.NewServeMux(
		// rawWebOption(jsonPb),
		httpBodyOption(),
		runtime.WithMetadata(metadata.ParseHeaderAndQueryToMD),
		runtime.WithErrorHandler(protoErrorHandler),
		// runtime.WithProtoErrorHandler(protoErrorHandler), // v1
	)

	return mux
}
