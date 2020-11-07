package service

import (
	"context"
	"github.com/pjoc-team/pay-gateway/pkg/service"
	pb "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/tracinggrpc"
	"google.golang.org/grpc"
)

type services struct {
	Discovery service.Discovery
}

func NewServices(Discovery service.Discovery){

}


func (s *services) InitDbClient(ctx context.Context) (pb.PayDatabaseServiceClient, error) {


	d, err := grpc.DialContext(
		ctx, "database-service:8080", grpc.WithChainUnaryInterceptor(
			tracinggrpc.
				TracingClientInterceptor(),
		),
	)
	if err != nil{
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
func (s *services) GetDatabaseService() (pb.PayDatabaseServiceClient, error) {
	return nil, nil
}
