package main

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/pjoc-team/pay-gateway/internal/service"
	"github.com/pjoc-team/pay-gateway/pkg/discovery"
	"github.com/pjoc-team/pay-gateway/pkg/util/db"
	pay "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"os"
	"time"
)

const serviceName = discovery.DatabaseService

var (
	c = &db.MysqlConfig{}
)

func flagSet() *pflag.FlagSet {
	set := pflag.NewFlagSet(serviceName.String(), pflag.ExitOnError)
	set.StringVar(
		&c.URL, "url",
		"root:111@tcp(127.0.0.1:3306)/pay_gateway?charset=utf8mb4&parseTime=true&loc=Local"+
			"&timeout=10s&collation=utf8mb4_unicode_ci", "mysql dsn",
	)
	set.IntVar(&c.MaxConn, "max-conn", 100, "max connection")
	set.DurationVarP(
		&c.MaxIdle, "max-idle", "w", 120*time.Second,
		"seconds of idle connection",
	)
	return set
}

func main() {
	log := logger.Log()

	validate := validator.New()
	err := validate.Struct(c)
	if err != nil {
		log.Fatalf("illegal configs, error: %v", err.Error())
	}

	if err != nil {
		log.Fatal(err.Error())
	}
	s, err := service.NewServer(serviceName.String())
	if err != nil {
		log.Fatal(err.Error())
	}
	set := flagSet()
	set.AddFlagSet(s.FlagSet)
	err = set.Parse(os.Args)
	if err != nil {
		panic(err.Error())
	}

	dbService, err := service.NewDatabaseService(s.Ctx, c)
	if err != nil {
		log.Fatal(err.Error())
	}
	grpcInfo := &service.GrpcInfo{
		RegisterGrpcFunc: func(ctx context.Context, server *grpc.Server) error {
			pay.RegisterPayDatabaseServiceServer(server, dbService)
			return nil
		},
		RegisterGatewayFunc: func(ctx context.Context, mux *runtime.ServeMux) error {
			err := pay.RegisterPayDatabaseServiceHandlerServer(ctx, mux, dbService)
			return err
		},
		Name: serviceName.String(),
	}
	s.Start(service.WithGrpc(grpcInfo))
}
