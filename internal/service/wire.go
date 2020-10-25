// +build wireinject

package service

//go:generate wire gen .

import (
	"github.com/google/wire"
	"github.com/pjoc-team/pay-gateway/pkg/configclient"
	"github.com/pjoc-team/pay-gateway/pkg/gateway"
	pay "github.com/pjoc-team/pay-proto/go"
)

var set = wire.NewSet(InitDbClient)

func NewPayGateway(configclients configclient.ConfigClients, clusterID string, concurrency int) (pay.PayGatewayServer, error) {
	wire.Build(set, gateway.NewPayGateway)
	return nil, nil
}
