package recover

import (
	"context"
	"net/http"
	"runtime/debug"

	"github.com/blademainer/commons/pkg/recoverable"
	"github.com/pjoc-team/pay-gateway/pkg/grpc/interceptors"
	"github.com/pjoc-team/tracing/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	interceptors.RegisterHttpInterceptor(recoverInterceptor)
}

// recoverInterceptor 感知panic错误
func recoverInterceptor(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			recoverable.WithRecoveryHandlerContext(
				r.Context(),
				func() {
					h.ServeHTTP(w, r)
				},
				CustomRecoverFunc,
			)
		},
	)
}

// CustomRecoverFunc recover func
func CustomRecoverFunc(ctx context.Context, p interface{}) (err error) {
	stack := debug.Stack()
	log := logger.ContextLog(ctx)
	log.Errorf("panic found: %v stack: %v", p, string(stack))
	return status.Errorf(codes.Unknown, "panic triggered: %v", p)
}
