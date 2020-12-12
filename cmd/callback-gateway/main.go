package main

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pjoc-team/pay-gateway/internal/service"
	"github.com/pjoc-team/pay-gateway/pkg/callback"
	"github.com/pjoc-team/pay-gateway/pkg/discovery"
	pay "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
)

const serviceName = discovery.ServiceName("callback-gateway")

func flagSet() *pflag.FlagSet {
	set := pflag.NewFlagSet(serviceName.String(), pflag.ExitOnError)
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

	err = s.Init(service.WithFlagSet(flagSet()))
	if err != nil {
		log.Fatal(err.Error())
	}

	server, err := callback.NewServer(s.GetDiscoveryServices())
	if err != nil {
		log.Fatalf("failed init server, error: %v", err.Error())
	}
	grpcInfo := &service.GrpcInfo{
		RegisterGrpcFunc: func(ctx context.Context, gs *grpc.Server) error {
			pay.RegisterChannelCallbackServer(gs, server)
			return nil
		},
		RegisterGatewayFunc: func(ctx context.Context, mux *runtime.ServeMux) error {
			err := pay.RegisterChannelCallbackHandlerServer(ctx, mux, server)
			return err
		},
		RegisterStreamFunc: pay.RegisterChannelCallbackHandler,
		Name:               serviceName.String(),
	}
	s.Start(service.WithGrpc(grpcInfo))
}
