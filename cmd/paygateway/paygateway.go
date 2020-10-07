package main

import (
	"context"
	"flag"
	"github.com/go-playground/validator/v10"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	wired "github.com/pjoc-team/pay-gateway/cmd"
	pay "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
	"google.golang.org/grpc"
)

type initConfig struct {
	configURI   string `validate:"url"`
	clusterID   string `validate:"required"`
	concurrency int    `validate:"gt=0"`
}

func init() {
	flag.StringVar(&c.configURI, "config-uri", "file://./pay-gateway.yaml", "pay gateway config uri")
	flag.StringVar(&c.clusterID, "cluster-id", "01", "cluster id for multiply cluster")
	flag.IntVar(&c.concurrency, "concurrency", 10000, "max concurrency order request per seconds")
}

var (
	c = &initConfig{}
)

func main() {
	flag.Parse()

	log := logger.Log()

	validate := validator.New()
	err := validate.Struct(c)
	if err != nil {
		log.Fatalf("illegal configs, error: %v", err.Error())
	}

	payGateway, err := wired.NewPayGateway(c.clusterID, c.concurrency)
	s, err := wired.NewServer("pay-gateway", &wired.GrpcInfo{
		RegisterGrpcFunc: func(ctx context.Context, server *grpc.Server) error {
			pay.RegisterPayGatewayServer(server, payGateway)
			return nil
		},
		RegisterGatewayFunc: func(ctx context.Context, mux *runtime.ServeMux) error {
			err := pay.RegisterPayGatewayHandlerServer(ctx, mux, payGateway)
			return err
		},
		Name: "pay-gateway",
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	s.Start()
}
