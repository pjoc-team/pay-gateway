package service

import (
	"context"
	"github.com/pjoc-team/pay-gateway/pkg/service"
	pb "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
	"github.com/pjoc-team/tracing/tracinggrpc"
	"google.golang.org/grpc"
)

type services struct {
	Discovery service.Discovery
}

func NewServices(discovery service.Discovery) *services {

}

func (s *services) InitDbClient(ctx context.Context) (pb.PayDatabaseServiceClient, error) {
	d, err := grpc.DialContext(
		ctx, "database-service:8080", grpc.WithChainUnaryInterceptor(
			tracinggrpc.
				TracingClientInterceptor(),
		),
	)
	if err != nil {
		return nil, err
	}
	client := pb.NewPayDatabaseServiceClient(d)
	return client, nil
}

// GetChannelClient get channel client of id
func (s *services) GetChannelClient(id string) (pb.PayChannelClient, error) {
	return nil, nil
}

// GetDatabaseService get channel client of id
func (s *services) GetDatabaseService(ctx context.Context) (pb.PayDatabaseServiceClient, error) {
	d, err := grpc.DialContext(
		ctx, "database-service:8080", grpc.WithChainUnaryInterceptor(
			tracinggrpc.
				TracingClientInterceptor(),
		),
	)
	if err != nil {
		return nil, err
	}
	client := pb.NewPayDatabaseServiceClient(d)
	return client, nil
}

func (s *services) initGrpc(
	ctx context.Context, serviceName string,
	grpcFunc func(conn *grpc.ClientConn) interface{},
) (interface{}, error) {
	log := logger.ContextLog(ctx)
	svc, err := s.Discovery.GetService(ctx, serviceName)
	if err != nil {
		log.Errorf("failed to get service: %v error: %v", serviceName, err.Error())
		return nil, err
	}

	target, err := svc.BuildTarget(ctx)
	if err != nil {
		log.Errorf("failed to build target of service: %v error: %v", serviceName, err.Error())
		return nil, err
	}

	d, err := grpc.DialContext(
		ctx, target, grpc.WithChainUnaryInterceptor(
			tracinggrpc.TracingClientInterceptor(),
		),
	)
	if err != nil {
		return nil, err
	}
	client := grpcFunc(d)
	return client, err
}
