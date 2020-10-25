package service

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
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

// RegisterGrpc 注册grpc服务的函数。请在该函数被调用时注册自己的grpc服务。
// 在服务启动时，会调用所有该类型函数。
type GrpcInfo struct {
	RegisterGrpcFunc    RegisterGrpcFunc
	RegisterGatewayFunc RegisterGatewayFunc

	Name string
}

// RegisterGrpc 注册grpc服务。这里会同时将grpc服务注册到grpc和gateway
func RegisterGrpc(name string, grpcFunc RegisterGrpcFunc, gatewayFunc RegisterGatewayFunc) {
	GrpcServices[name] = GrpcInfo{
		Name:                name,
		RegisterGrpcFunc:    grpcFunc,
		RegisterGatewayFunc: gatewayFunc,
	}
}
