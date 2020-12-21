package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/blademainer/commons/pkg/benchmark"
	"github.com/pjoc-team/pay-gateway/pkg/configclient"
	"github.com/pjoc-team/pay-gateway/pkg/discovery"
	"github.com/pjoc-team/pay-gateway/pkg/generator"
	"github.com/pjoc-team/pay-gateway/pkg/sign"
	pb "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
	"google.golang.org/protobuf/types/known/timestamppb"
	"sync"
)

var (
	concurrent = flag.Int("concurrent", 1, "concurrent")
	count      = flag.Int("count", 1, "count per thread")
)

func main() {
	err := logger.MinReportCallerLevel(logger.DebugLevel)
	if err != nil {
		panic(err)
	}

	log := logger.Log()

	g := generator.New("T1", 1_000_000)
	g.Debug()
	id := g.GenerateID()
	fmt.Println(id)

	errOrderFailed := errors.New("order failed")

	merchantConfig := &configclient.MerchantConfig{
		AppID:               "1",
		GatewayRSAPublicKey: "",
		GatewayRSAPrivateKey: `MIIEowIBAAKCAQEA2KaaJp7JeW91WlQCfZeS14US/ot9hIJViutv3JHojdgTx+8A
8psStKaPl2Ac/MTJ/3mHeopCObmgjw/Au/Ne0PS1rveY0Pcazwnp+R1TDP2H9jag
c3GJWS6cvHLB/B4uP3LOnPXN8ctwDVsF19b/howVKUKX6RAX7R2VAEyTIZJIEIQE
0fNvRCWqbVv1RB3LU4cbQmW6nX8dP793fP8s/Lhzcj6vS6UKxLVl5CrCCGIJIBYc
1mI8RbUYvGqwiONEnEwYvOioAoAlkMIXdFndIjngHe7JYfGW1NtPzHLG5yw8anYT
D/3du7hJ/kSN0WM6NLa0P/vbR5+mxVdoRzY+kQIDAQABAoIBABaE2qkBADgbGbuV
19xuENlN/7dtkFJhqbqS1kG6+M0llIjHkvWkoMEePvahCuJLIiPn4ekezdtqLAIy
xPnERiq6BNh26+9sf+DdSvCV17gV8jfpXawiNQCME8aStw8Zo/z8VfWCpzFmz/LT
bzwMIOs/TEPJpDiZb6M52+74BqMKfHTY14YOF8Xr4fiaUFpNTViHeOQXKzoG5PF4
GLlhg7YNgEnjyc578izCoFp/xTjBBHQ7dtu+EnzmXD9QTlz7xUYt4P2TjUEBKy1o
xSxDpgFL+BKYgRazilkrJ2hesbGCvbxDzcd4ivzpfmvqkN74Lq0vF9voL1JSd6D2
3l/R9bECgYEA7nbQgsK3ResReUMumJvE4y1sl2D+rt24QHlu+jOJqVkpAo0L4HLj
vCX0Y8tBfG/hDc5iC12YILCn+EEb9bD2giURg7V+cA+K4IJrbLTbnna/UlvA1PFK
3kHFosdCk5cRlpAppEBLQEUjlf7mjp6k2Xxy71ozg4KlB3wf3QCrs2MCgYEA6JUk
iXd/lntdjb7V/QdwhVFdp/lzst0ClE4q04RNL8ZjwmSrYOAOGO2ktKOBG8lGT3P6
54/BASn9TMOXks8gPE3r/pN+21RGvOq2xtHNOrnV5g1RvlqHtwtv2RUxoEoTKPjB
m6KDeLrPNCuGZ3bzYUUNAys66v3iWM5PK2s1GnsCgYEAjtKyx95/jmzgJlTKj7Sc
E8SdCX2ajHlXZaZVhZ1gkgFIwrJfrqqhI4tH+I1AR5tqm65EorIH72xe7h1w9ZJr
0j8JYm1NsShd8WGrnYwlDZ/prxYtRFzQjpWuHXRit6r/acImbq3jZDcEvU3SIRF7
gpc674iC2f1hgj4hh2hjbikCgYBWfG8ztv34xTMKrHYCOyv6R0FeXwJI9qoo39BJ
Cx9wroMWHD0mLurPFj9y9IHkBTph/SzFwszwU97fFrRcYS0Jf6hL6Cj6AiKzyUvi
Ls30EnqZq0ZEVIG27UfQH3NuuVzalXXZG9trn3vBWJYID1F9UCIAlai5DWOHxl/m
M11x1QKBgCvi33kFLll6SVIKfkwt1Hja/DGlyq/M/xN4qn/wQwGKYzzIU+73SbQR
g44GAiYVMJQrjISg/RVd4ClDxZ+A0cpumfpuSJdcT210L4u5FkuTQAmLZ2HOhTzK
mW9/iR9koFHtTzTKhhYIgSWy9EWkQmcyrOKnEPYqMJjMobDJ1AuG`,
		MerchantRSAPublicKey: "",
		Md5Key:               "",
	}

	b := benchmark.New(
		*concurrent, *count, func(ctx context.Context) error {
			gc, err := discovery.DialTarget(ctx, "http://127.0.0.1:9090")
			if err != nil {
				log.Fatal(err.Error())
			}
			defer gc.Close()
			client := pb.NewPayGatewayClient(gc)

			request := &pb.PayRequest{}
			request.AppId = "1"
			request.OutTradeNo = g.GenerateID()
			log.Infof(request.OutTradeNo)
			request.Method = pb.Method_WEB
			request.OrderTime = timestamppb.Now()
			request.ChannelId = "mock"
			request.ProductName = "Apple 12"
			request.ProductDescribe = "Hello jobs"
			request.UserIp = "127.0.0.1"
			request.PayAmount = 1

			signValidator := sign.NewCheckSignValidator()
			paramsString := signValidator.ParamsCompacter.ParamsToString(request)
			log := logger.ContextLog(ctx)
			if log.IsDebugEnabled() {
				log.Debugf("Build interface: %v to string: %v", request, paramsString)
			}
			signString, err := sign.GenerateSign(
				ctx, "utf-8", paramsString, merchantConfig, sign.TypeSha256WithRSA,
			)
			if err != nil {
				log.Error(err.Error())
				return err
			}
			request.Sign = signString
			request.SignType = string(sign.TypeSha256WithRSA)

			response, err := client.Pay(ctx, request)
			if err != nil {
				log.Error(err.Error())
				return err
			}
			if response.Result.Code != pb.ReturnResultCode_CODE_SUCCESS {
				log.Errorf("result: %v", response)
				return errOrderFailed
			}

			return nil
		},
	)

	benchmarks := []*benchmark.BenchMark{b}

	rootContext, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}
	for _, b := range benchmarks {
		wg.Add(1)
		b := b
		go func() {
			result := b.Start(rootContext)
			fmt.Printf("bench result: %#v \n", result)
			wg.Done()
		}()
	}

	wg.Wait()
}
