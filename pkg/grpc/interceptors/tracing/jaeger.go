package tracing

import (
	"context"
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pjoc-team/pay-gateway/pkg/grpc/interceptors"
	"github.com/pjoc-team/tracing/logger"
	"github.com/pjoc-team/tracing/tracing"
	"github.com/pjoc-team/tracing/util"
)

// TraceID http响应header内返回的traceID
const TraceID = "trace-id"

func init() {
	interceptors.RegisterHttpInterceptor(ServerInterceptor)
}

// ServerInterceptor 拦截grpc gateway生成tracing信息
func ServerInterceptor(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			log := logger.Log()
			newCtx := r.Context()
			spanCtx, err := opentracing.GlobalTracer().Extract(
				opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header),
			)
			if err != nil && err != opentracing.ErrSpanContextNotFound {
				log.Errorf("extract from header err: %v", err)
			} else {
				span := opentracing.GlobalTracer().StartSpan(
					r.RequestURI, ext.RPCServerOption(spanCtx),
				)
				defer span.Finish()
				newCtx = opentracing.ContextWithSpan(newCtx, span)
				requestID := r.Header.Get(string(tracing.HttpHeaderKeyXRequestID))
				if requestID != "" {
					span.SetTag(string(tracing.SpanTagKeyHttpRequestID), requestID)
					newCtx = context.WithValue(newCtx, tracing.SpanTagKeyHttpRequestID, requestID)
					w.Header().Add(string(tracing.HttpHeaderKeyXRequestID), requestID)
				}
				w.Header().Add(TraceID, util.GetTraceID(newCtx))
			}
			h.ServeHTTP(w, r.WithContext(newCtx))
		},
	)
}
