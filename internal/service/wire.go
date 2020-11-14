// +build wireinject

package service

//go:generate wire gen .

import (
	"github.com/google/wire"
	"github.com/pjoc-team/pay-gateway/pkg/configclient"
	"github.com/pjoc-team/pay-gateway/pkg/discovery"
	"github.com/pjoc-team/pay-gateway/pkg/gateway"
	pay "github.com/pjoc-team/pay-proto/go"
)

// NewPayGateway create pay gateway service
func NewPayGateway(
	configclients configclient.ConfigClients, clusterID string, concurrency int,
	services *discovery.Services,
) (pay.PayGatewayServer, error) {
	wire.Build(gateway.NewPayGateway)
	return nil, nil
}
