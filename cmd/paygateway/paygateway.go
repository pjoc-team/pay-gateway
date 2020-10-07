package main

import (
	"flag"
	"github.com/go-playground/validator/v10"
	"github.com/pjoc-team/pay-gateway/pkg/config"
	"github.com/pjoc-team/pay-gateway/pkg/gateway"
	tracinglogger "github.com/pjoc-team/tracing/logger"
	"google.golang.org/grpc"
	"math/rand"
	"time"
)

type initConfig struct {
	configURI string `validate:"url"`
	clusterID string `validate:"required"`
	concurrency int `validate:"gt=0"`
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
	rand.Seed(int64(time.Now().Nanosecond()))

	log := tracinglogger.Log()

	validate := validator.New()
	err := validate.Struct(c)
	if err != nil {
		log.Fatalf("illegal configs, error: %v", err.Error())
	}

	server, err := config.InitConfigServer(c.configURI)

	payGateway, err := gateway.NewPayGateway(server, c.clusterID, c.concurrency)


}
