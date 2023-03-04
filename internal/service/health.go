package service

import (
	"encoding/json"
	"net/http"

	"github.com/pjoc-team/tracing/logger"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// healthInterceptor 拦截health请求
func healthInterceptor(healthServer *health.Server) http.Handler {
	return http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			log := logger.ContextLog(request.Context())
			checkRequest := &healthpb.HealthCheckRequest{}
			check, err3 := healthServer.Check(request.Context(), checkRequest)
			if err3 != nil {
				log.Errorf("unhealth status")
				writer.WriteHeader(http.StatusBadGateway)
				return
			}
			marshal, err3 := json.Marshal(check)
			if err3 != nil {
				log.Errorf("failed to marshal HealthResponse: %v", err3.Error())
				writer.WriteHeader(http.StatusBadGateway)
				return
			}
			_, err3 = writer.Write(marshal)
			if err3 != nil {
				log.Errorf("failed to write Response: %v error: %v", string(marshal), err3.Error())
				writer.WriteHeader(http.StatusBadGateway)
				return
			}
		},
	)
}
