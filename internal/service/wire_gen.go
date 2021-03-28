// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package service

import (
	"context"
	"github.com/google/wire"
	"github.com/pjoc-team/pay-gateway/pkg/configclient"
	"github.com/pjoc-team/pay-gateway/pkg/dbservice"
	"github.com/pjoc-team/pay-gateway/pkg/discovery"
	"github.com/pjoc-team/pay-gateway/pkg/gateway"
	"github.com/pjoc-team/pay-gateway/pkg/util/db"
	"github.com/pjoc-team/pay-proto/go"
)

import (
	_ "github.com/pjoc-team/pay-gateway/pkg/config/backend/file"
	_ "net/http/pprof"
)

// Injectors from wire.go:

func NewPayGateway(configclients configclient.ConfigClients, clusterID string, concurrency int, services *discovery.Services) (pay.PayGatewayServer, error) {
	payGatewayServer, err := gateway.NewPayGateway(configclients, clusterID, concurrency, services)
	if err != nil {
		return nil, err
	}
	return payGatewayServer, nil
}

func NewDatabaseService(ctx context.Context, config *db.MysqlConfig) (pay.PayDatabaseServiceServer, error) {
	gormDB, err := db.InitDb(ctx, config)
	if err != nil {
		return nil, err
	}
	payDatabaseServiceServer, err := dbservice.NewServer(gormDB)
	if err != nil {
		return nil, err
	}
	return payDatabaseServiceServer, nil
}

// wire.go:

var set = wire.NewSet(db.InitDb)
