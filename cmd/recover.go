package wired

import (
	"context"
	"github.com/pjoc-team/tracing/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"runtime/debug"
)

func customRecoverFunc(ctx context.Context, p interface{}) (err error) {
	stack := debug.Stack()
	log := logger.ContextLog(ctx)
	log.Errorf("panic found: %v stack: %v", p, string(stack))
	return status.Errorf(codes.Unknown, "panic triggered: %v", p)
}
