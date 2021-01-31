package discovery

import (
	"context"
	pb "github.com/pjoc-team/pay-proto/go"
	"google.golang.org/grpc"
)

const (
	// DatabaseService db service
	DatabaseService ServiceName = "database-service"
	// PayGateway pay gateway
	PayGateway ServiceName = "pay-gateway"
	// Settlement settlement gateway
	Settlement ServiceName = "settlement"
)

const (
	// ChannelPrefix channel service name prefix
	ChannelPrefix = "channel-"
)

// GetSettlementClient settlement client
func (s *Services) GetSettlementClient(ctx context.Context) (
	client pb.SettlementGatewayClient,
	pubBackClientFunc PutBackClientFunc, err error,
) {
	_, pubBackClientFunc, err = s.initGrpc(
		ctx, Settlement.String(), func(conn *grpc.ClientConn) interface{} {
			client = pb.NewSettlementGatewayClient(conn)
			return client
		},
	)
	return
}

// GetChannelClient get channel client of id
func (s *Services) GetChannelClient(ctx context.Context, id string) (
	client pb.PayChannelClient, pubBackClientFunc PutBackClientFunc,
	err error,
) {
	_, pubBackClientFunc, err = s.initGrpc(
		ctx,
		ChannelPrefix + id,
		func(conn *grpc.ClientConn) interface{} {
			client = pb.NewPayChannelClient(conn)
			return client
		},
	)
	return
}

// GetDatabaseService get channel client of id
func (s *Services) GetDatabaseService(ctx context.Context) (
	pb.PayDatabaseServiceClient, PutBackClientFunc,
	error,
) {
	var client pb.PayDatabaseServiceClient
	_, putBackFunc, err := s.initGrpc(
		ctx, DatabaseService.String(), func(conn *grpc.ClientConn) interface{} {
			client = pb.NewPayDatabaseServiceClient(conn)
			return client
		},
	)
	return client, putBackFunc, err
}
