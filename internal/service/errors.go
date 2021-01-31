package service

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pjoc-team/tracing/logger"
	"net/http"
)

// errorHandler 接收错误
func errorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, writer http.ResponseWriter, request *http.Request, err error) {
	log := logger.ContextLog(ctx)
	log.Errorf("proto error: %v request: %#v", err.Error(), request)
	marshal, err2 := marshaler.Marshal(err)
	if err2 != nil {
		log.Errorf("failed to marshal e: %v request: %#v error: %v", err.Error(), request, err2.Error())
		return
	}
	_, err3 := writer.Write(marshal)
	if err3 != nil {
		log.Errorf("failed to write e: %v request: %#v error: %v", err.Error(), request, err3.Error())
		return
	}
}
