package http

import (
	"context"
	"io"
	"net/http"

	"github.com/pjoc-team/pay-gateway/pkg/grpc/interceptors"
	"github.com/pjoc-team/tracing/logger"
)

const (
	// ContextHttpRequestBody context key
	ContextHttpRequestBody interceptors.InterceptorContextKey = "request-body"

	// ContextHttpRequestMethod context key
	ContextHttpRequestMethod interceptors.InterceptorContextKey = "request-method"
)

func init() {
	interceptors.RegisterHttpInterceptor(httpBodyInterceptor)
}

// httpBodyInterceptor 拦截health请求
func httpBodyInterceptor(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := logger.ContextLog(ctx)
			body, err := io.ReadAll(r.Body)
			if err != nil {
				log.Errorf("failed to read http body, err: %v", err.Error())
				return
			}
			newCtx := context.WithValue(r.Context(), ContextHttpRequestBody, body)
			newCtx = context.WithValue(newCtx, ContextHttpRequestMethod, r.Method)
			h.ServeHTTP(w, r.WithContext(newCtx))
		},
	)
}
