package interceptors

import (
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// InterceptorContextKey write to go context
type InterceptorContextKey string

// Interceptor http interceptor
type Interceptor func(http.Handler) http.Handler

var httpInterceptors []Interceptor

// RegisterHttpInterceptor register http interceptor
func RegisterHttpInterceptor(i Interceptor) {
	httpInterceptors = append(httpInterceptors, i)
}

// Intercept ServeMux
func Intercept(mux *runtime.ServeMux) http.Handler {
	var h http.Handler = mux
	for _, hi := range httpInterceptors {
		h = hi(h)
	}
	return h
}
