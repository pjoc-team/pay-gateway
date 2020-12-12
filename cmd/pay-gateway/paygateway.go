package main

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pjoc-team/pay-gateway/internal/service"
	"github.com/pjoc-team/pay-gateway/pkg/configclient"
	"github.com/pjoc-team/pay-gateway/pkg/discovery"
	pay "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"os"
)

const serviceName = discovery.PayGateway

var (
	c = &initConfig{}
)

type initConfig struct {
	clusterID string `validate:"required"`
	concurrency int  `validate:"gt=0"`
}

func flagSet() *pflag.FlagSet {
	set := pflag.NewFlagSet("pay-gateway", pflag.ExitOnError)
	set.StringVar(&c.clusterID, "cluster-id", "01", "cluster id for multiply cluster")
	set.IntVar(&c.concurrency, "concurrency", 1_000_000,
		"max concurrency order request per seconds")
	return set
}

func main() {
	log := logger.Log()

	validate := validator.New()
	err := validate.Struct(c)
	if err != nil {
		log.Fatalf("illegal configs, error: %v", err.Error())
	}

	configClients, configFlagSet, err := configclient.NewConfigClients(
		configclient.WithMerchantConfigServer(true),
		configclient.WithAppIDChannelConfigServer(true),
		configclient.WithPayConfigServer(true),
	)
	if err != nil {
		log.Fatal(err.Error())
	}
	set := flagSet()
	set.AddFlagSet(configFlagSet)
	err = set.Parse(os.Args)
	if err != nil {
		panic(err.Error())
	}
	s, err := service.NewServer(serviceName.String())
	if err != nil {
		log.Fatal(err.Error())
	}

	payGateway, err := service.NewPayGateway(
		configClients, c.clusterID, c.concurrency, s.GetDiscoveryServices(),
	)
	if err != nil {
		log.Fatal(err.Error())
	}
	grpcInfo := &service.GrpcInfo{
		RegisterGrpcFunc: func(ctx context.Context, server *grpc.Server) error {
			pay.RegisterPayGatewayServer(server, payGateway)
			return nil
		},
		RegisterGatewayFunc: func(ctx context.Context, mux *runtime.ServeMux) error {
			err := pay.RegisterPayGatewayHandlerServer(ctx, mux, payGateway)
			return err
		},
		Name: serviceName.String(),
	}
	s.Start(service.WithGrpc(grpcInfo), service.WithFlagSet(set))
}
