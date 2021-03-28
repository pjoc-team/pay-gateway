// +build wireinject

package service

//go:generate go run github.com/google/wire/cmd/wire gen .

import (
	"context"
	"github.com/google/wire"
	"github.com/pjoc-team/pay-gateway/pkg/configclient"
	"github.com/pjoc-team/pay-gateway/pkg/discovery"
	"github.com/pjoc-team/pay-gateway/pkg/gateway"
	"github.com/pjoc-team/pay-gateway/pkg/dbservice"
	"github.com/pjoc-team/pay-gateway/pkg/util/db"
	pay "github.com/pjoc-team/pay-proto/go"
)

var set = wire.NewSet(db.InitDb)

// NewPayGateway create pay gateway service
func NewPayGateway(
	configclients configclient.ConfigClients, clusterID string, concurrency int,
	services *discovery.Services,
) (pay.PayGatewayServer, error) {
	wire.Build(gateway.NewPayGateway)
	return nil, nil
}

// NewPayGateway create pay gateway service
func NewDatabaseService(
	ctx context.Context, config *db.MysqlConfig,
) (pay.PayDatabaseServiceServer, error) {
	wire.Build(set, dbservice.NewServer)
	return nil, nil
}
