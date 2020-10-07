// +build wireinject

package wired

//go:generate wire gen .

import (
	"github.com/google/wire"
	"github.com/pjoc-team/pay-gateway/pkg/gateway"
	pay "github.com/pjoc-team/pay-proto/go"
)

var set = wire.NewSet(InitConfigClients, InitDbClient)

func NewPayGateway(clusterID string, concurrency int) (pay.PayGatewayServer, error) {
	wire.Build(set, gateway.NewPayGateway)
	return nil, nil
}
