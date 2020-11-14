package main

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/pjoc-team/pay-gateway/internal/service"
	"github.com/pjoc-team/pay-gateway/pkg/channels/mock"
	"github.com/pjoc-team/pay-gateway/pkg/config"
	_ "github.com/pjoc-team/pay-gateway/pkg/config/file"
	"github.com/pjoc-team/pay-gateway/pkg/discovery"
	pay "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"os"
)

const serviceName = discovery.ServiceName("mock")

var (
	configURL string
)

func flagSet() *pflag.FlagSet {
	set := pflag.NewFlagSet(serviceName.String(), pflag.ExitOnError)
	set.StringVarP(
		&configURL, "config-url", "c",
		"file://./conf/biz/channel/mock.yaml", "config url",
	)
	return set
}

func main() {
	log := logger.Log()

	s, err := service.NewServer(serviceName.String())
	if err != nil {
		log.Fatal(err.Error())
	}

	if err != nil {
		log.Fatal(err.Error())
	}

	set := flagSet()
	set.AddFlagSet(s.FlagSet)
	err = set.Parse(os.Args)
	if err != nil {
		panic(err.Error())
	}

	if configURL == "" {
		log.Fatal("config url is nill")
	}

	cs, err := config.InitConfigServer(configURL)
	if err != nil {
		log.Fatalf("illegal configs, error: %v", err.Error())
	}

	server, err := mock.NewServer(cs)
	grpcInfo := &service.GrpcInfo{
		RegisterGrpcFunc: func(ctx context.Context, gs *grpc.Server) error {
			pay.RegisterPayChannelServer(gs, server)
			return nil
		},
		RegisterGatewayFunc: func(ctx context.Context, mux *runtime.ServeMux) error {
			err := pay.RegisterPayChannelHandlerServer(ctx, mux, server)
			return err
		},
		Name: serviceName.String(),
	}
	s.Start(service.WithGrpc(grpcInfo))
}
