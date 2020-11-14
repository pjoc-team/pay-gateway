package main

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pjoc-team/pay-gateway/internal/service"
	"github.com/pjoc-team/pay-gateway/pkg/configclient"
	pay "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"os"
)

const serviceName = "pay-gateway"

var (
	c = &MysqlConfig{}
)

// MysqlConfig mysql配置
type MysqlConfig struct {
	URL     string `yaml:"url" json:"url"`
	MaxConn int    `yaml:"max_conn" json:"max_conn"`
	MaxIdle int    `yaml:"max_idle" json:"max_idle"`
}

func flagSet() *pflag.FlagSet {
	set := pflag.NewFlagSet("database-service", pflag.ExitOnError)
	set.StringVar(&c.clusterID, "cluster-id", "01", "cluster id for multiply cluster")
	set.IntVar(&c.concurrency, "concurrency", 10000, "max concurrency order request per seconds")
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
	)
	if err != nil {
		log.Fatal(err.Error())
	}
	s, fs, err := service.NewServer(serviceName)
	if err != nil {
		log.Fatal(err.Error())
	}
	set := flagSet()
	set.AddFlagSet(configFlagSet)
	set.AddFlagSet(fs)
	err = set.Parse(os.Args)
	if err != nil {
		panic(err.Error())
	}

	payGateway, err := service.NewPayGateway(configClients, c.clusterID, c.concurrency, s.GetServices())
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
		Name: serviceName,
	}
	s.Start(service.WithGrpc(grpcInfo))
}
