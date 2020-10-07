package service

import (
	"context"
	"encoding/json"
	"github.com/blademainer/commons/pkg/recoverable"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	tracinglogger "github.com/pjoc-team/tracing/logger"
	"github.com/pjoc-team/tracing/tracing"
	"github.com/pjoc-team/tracing/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"net/http"
)

// TraceID http响应header内返回的traceID
const TraceID = "trace-id"

type interceptor func(http.Handler) http.Handler

var httpInterceptors = []interceptor{recoverInterceptor, tracingServerInterceptor}

//recoverInterceptor 感知panic错误
func recoverInterceptor(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recoverable.WithRecoveryHandlerContext(
			r.Context(),
			func() {
				h.ServeHTTP(w, r)
			},
			func(ctx context.Context, p interface{}) (err error) {
				return status.Errorf(codes.Unknown, "panic triggered: %v", p)
			})
	})
}

//tracingServerInterceptor 拦截grpc gateway生成tracing信息
func tracingServerInterceptor(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := tracinglogger.ContextLog(nil)
		newCtx := r.Context()
		spanCtx, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		if err != nil && err != opentracing.ErrSpanContextNotFound {
			logger.Errorf("extract from header err: %v", err)
		} else {
			span := opentracing.GlobalTracer().StartSpan(r.RequestURI, ext.RPCServerOption(spanCtx))
			defer span.Finish()
			newCtx = opentracing.ContextWithSpan(newCtx, span)
			requestID := r.Header.Get(tracing.HttpHeaderKeyXRequestID)
			if requestID != "" {
				span.SetTag(tracing.SpanTagKeyHttpRequestID, requestID)
				newCtx = context.WithValue(newCtx, tracing.SpanTagKeyHttpRequestID, requestID)
				w.Header().Add(tracing.HttpHeaderKeyXRequestID, requestID)
			}
			w.Header().Add(TraceID, util.GetTraceID(newCtx))
		}
		h.ServeHTTP(w, r.WithContext(newCtx))
	})
}

//healthInterceptor 拦截health请求
func healthInterceptor(healthServer *health.Server) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		logger := tracinglogger.ContextLog(request.Context())
		checkRequest := &healthpb.HealthCheckRequest{}
		check, err3 := healthServer.Check(request.Context(), checkRequest)
		if err3 != nil {
			logger.Errorf("unhealth status")
			writer.WriteHeader(http.StatusBadGateway)
			return
		}
		marshal, err3 := json.Marshal(check)
		if err3 != nil {
			logger.Errorf("failed to marshal HealthResponse: %v", err3.Error())
			writer.WriteHeader(http.StatusBadGateway)
			return
		}
		_, err3 = writer.Write(marshal)
		if err3 != nil {
			logger.Errorf("failed to write Response: %v error: %v", string(marshal), err3.Error())
			writer.WriteHeader(http.StatusBadGateway)
			return
		}
	})
}
