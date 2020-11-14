package discovery

import (
	"context"
	grpcdialer "github.com/blademainer/commons/pkg/grpc"
	pb "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
	"github.com/pjoc-team/tracing/tracinggrpc"
	"github.com/prometheus/common/log"
	"google.golang.org/grpc"
	"net/url"
)

// ServiceName service name
type ServiceName string

const (
	// DatabaseService db service
	DatabaseService ServiceName = "database-service"
	// PayGateway pay gateway
	PayGateway ServiceName = "pay-gateway"
)

func (s ServiceName) String() string {
	return string(s)
}

// Services defined services
type Services struct {
	Discovery *Discovery
}

// NewServices create service discovery
func NewServices(discovery *Discovery) *Services {
	s := &Services{
		Discovery: discovery,
	}
	return s
}

// GetChannelClient get channel client of id
func (s *Services) GetChannelClient(ctx context.Context, id string) (
	client pb.PayChannelClient,
	err error,
) {
	_, err = s.initGrpc(
		ctx, id, func(conn *grpc.ClientConn) interface{} {
			client = pb.NewPayChannelClient(conn)
			return client
		},
	)
	return
}

// GetDatabaseService get channel client of id
func (s *Services) GetDatabaseService(ctx context.Context) (pb.PayDatabaseServiceClient, error) {
	var client pb.PayDatabaseServiceClient
	_, err := s.initGrpc(
		ctx, DatabaseService.String(), func(conn *grpc.ClientConn) interface{} {
			client = pb.NewPayDatabaseServiceClient(conn)
			return client
		},
	)
	return client, err
}

func (s *Services) initGrpc(
	ctx context.Context, serviceName string,
	grpcFunc func(conn *grpc.ClientConn) interface{},
) (interface{}, error) {
	log := logger.ContextLog(ctx)
	svc, err := s.Discovery.GetService(ctx, serviceName)
	if err != nil {
		log.Errorf("failed to get service: %v error: %v", serviceName, err.Error())
		return nil, err
	}
	if svc.Protocol != "" && svc.Protocol != GRPC {
		log.Errorf(
			"service: %v's protocol is not grpc, "+
				"actual is: %v and continue try to connect by grpc protocol",
			serviceName,
			svc.Protocol,
		)
	}

	target, err := svc.BuildTarget(ctx)
	if err != nil {
		log.Errorf("failed to build target of service: %v error: %v", serviceName, err.Error())
		return nil, err
	}

	d, err := DialTarget(ctx, target)

	if err != nil {
		return nil, err
	}
	client := grpcFunc(d)
	return client, err
}

// DialTarget dial grpc target
func DialTarget(ctx context.Context, target string) (*grpc.ClientConn, error) {
	u, err := url.Parse(target)
	if err != nil {
		log.Errorf(
			"failed to build target: %v, error: %v", target, err.Error(),
		)
		return nil, err
	}
	d, err := grpcdialer.DialUrl(
		ctx, *u, grpc.WithChainUnaryInterceptor(
			tracinggrpc.TracingClientInterceptor(),
		),
	)
	return d, err
}
