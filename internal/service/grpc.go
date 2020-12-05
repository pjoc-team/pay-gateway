package service

import (
	"context"
	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"net/http"
	"strings"
)

var (
	// GrpcServices 需要注册grpc的服务
	GrpcServices = make(map[string]GrpcInfo)
)

// RegisterGrpcFunc 注册grpc函数
type RegisterGrpcFunc func(ctx context.Context, server *grpc.Server) error

// RegisterGatewayFunc 注册gatway函数
type RegisterGatewayFunc func(ctx context.Context, mux *runtime.ServeMux) error

// GrpcInfo 注册grpc服务的函数。请在该函数被调用时注册自己的grpc服务。
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

// allowCORS allows Cross Origin Resoruce Sharing from any origin.
// Don't do this without consideration in production systems.
func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				preflightHandler(w, r)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

// preflightHandler adds the necessary headers in order to serve
// CORS from any origin using the methods "GET", "HEAD", "POST", "PUT", "DELETE"
// We insist, don't do this without consideration in production systems.
func preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept", "Authorization"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
	glog.Infof("preflight request for %s", r.URL.Path)
}