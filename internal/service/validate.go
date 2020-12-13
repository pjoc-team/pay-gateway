package service

import (
	"context"
	"google.golang.org/grpc"
)

// ValidatorInterceptor envoy validator interceptor
func ValidatorInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		if v, ok := req.(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return nil, err
			}
		}

		resp, err = handler(ctx, req)
		return resp, err
	}
}
