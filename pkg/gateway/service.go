package gateway

import (
	"flag"
	"fmt"
	"github.com/pjoc-team/base-service/pkg/generator"
	gc "github.com/pjoc-team/base-service/pkg/grpc"
	"github.com/pjoc-team/base-service/pkg/logger"
	"github.com/pjoc-team/base-service/pkg/model"
	"github.com/pjoc-team/base-service/pkg/recover"
	"github.com/pjoc-team/base-service/pkg/service"
	"github.com/pjoc-team/pay-gateway/pkg/validator"
	pb "github.com/pjoc-team/pay-proto/go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const ETCD_DIR_ROOT = "/pub/pjoc/pay"

type PayGatewayService struct {
	*service.GatewayConfig
	OrderGenerator *generator.OrderGenerator
	*service.Service
	*gc.GrpcClientFactory
}

type RequestContext struct {
	GatewayOrderId     string
	ChannelAccount     string
	PayRequest         *pb.PayRequest
	PayOrder           *pb.PayOrder
	ChannelPayRequest  *pb.ChannelPayRequest
	ChannelPayResponse *pb.ChannelPayResponse
	err                error
}

func BuildParamsErrorResponse(err error) *pb.PayResponse {
	response := &pb.PayResponse{}
	response.Result = &pb.ReturnResult{Code: pb.ReturnResultCode_CODE_PARAMS_ERROR, Message: "PARAMS_ERROR", Describe: err.Error()}
	return response
}
func BuildSystemErrorResponse(err error) *pb.PayResponse {
	response := &pb.PayResponse{}
	response.Result = &pb.ReturnResult{Code: pb.ReturnResultCode_CODE_SYSTEM_ERROR, Message: "SYSTEM_ERROR", Describe: err.Error()}
	return response
}

func (svc *PayGatewayService) Pay(ctx context.Context, request *pb.PayRequest) (response *pb.PayResponse, err error) {
	defer recover.Recover()
	logger.Log.Debugf("New request: %v", request)
	if err = request.Validate(); err != nil {
		return BuildParamsErrorResponse(err), nil
	}
	if err = validator.Validate(*request, *svc.GatewayConfig); err != nil {
		return BuildParamsErrorResponse(err), nil
	}
	var cfg model.AppIdChannelConfig
	if cfg, err = svc.processChannelIdIfNotPresent(request); err != nil {
		err = fmt.Errorf("could'nt found config of channelId: %v", request.ChannelId)
		return BuildParamsErrorResponse(err), nil
	}

	response = &pb.PayResponse{}
	requestContext := &RequestContext{}

	gatewayOrderId := svc.OrderGenerator.GenerateOrderId()
	requestContext.GatewayOrderId = gatewayOrderId
	requestContext.PayRequest = request
	requestContext.ChannelAccount = cfg.ChannelAccount

	if result, e := svc.SavePayOrder(requestContext); e != nil || result.Code != pb.ReturnResultCode_CODE_SUCCESS {
		err = e
		response.Result = result
		return BuildSystemErrorResponse(err), nil
	}

	var client pb.PayChannelClient
	client, err = svc.GetChannelClient(request.GetChannelId())
	if client == nil || err != nil {
		logger.Log.Errorf("Failed to get channelClient! channelId: %s, error: %s, ", request.GetChannelId(), err.Error())
		return BuildSystemErrorResponse(err), nil
	} else {
		logger.Log.Debugf("Got client: %v for channelId: %s", client, request.GetChannelId())
	}
	var channelPayRequest *pb.ChannelPayRequest
	if channelPayRequest, err = svc.GenerateChannelPayRequest(requestContext); err != nil {
		return BuildSystemErrorResponse(err), nil
	}
	var channelPayResponse *pb.ChannelPayResponse
	if channelPayResponse, err = client.Pay(ctx, channelPayRequest); err != nil {
		logger.Log.Errorf("Pay channel failed! err: %s channelPayResponse: %v", err.Error(), channelPayResponse)
		requestContext.err = err
		svc.UpdatePayOrder(requestContext)
		return BuildSystemErrorResponse(err), nil
	} else if channelPayResponse == nil || channelPayResponse.Data == nil {
		err = fmt.Errorf("channel response fail! response: %v", channelPayResponse)
		logger.Log.Errorf(err.Error())
		return BuildSystemErrorResponse(err), nil
	}
	requestContext.ChannelPayResponse = channelPayResponse
	if result, e := svc.UpdatePayOrder(requestContext); e != nil || result.Code != pb.ReturnResultCode_CODE_SUCCESS {
		response.Result = result
		err = e
		return
	}

	response.Data = channelPayResponse.Data
	response.GatewayOrderId = gatewayOrderId
	response.Result = SUCCESS_RESULT

	return response, nil
}

// 如果没有传入channelId，则根据method找可用的channelId
func (svc *PayGatewayService) processChannelIdIfNotPresent(request *pb.PayRequest) (channelConfig model.AppIdChannelConfig, err error) {
	if svc.AppIdAndChannelConfigMap == nil {
		err = fmt.Errorf("failed to found info of appId: %v", request.AppId)
		return
	}
	config := (*svc.AppIdAndChannelConfigMap)[request.AppId]
	if config.ChannelConfigs == nil {
		err = fmt.Errorf("failed to found info of appId: %v", request.AppId)
		return
	}
	for _, config := range config.ChannelConfigs {
		found := (request.ChannelId != "" && request.ChannelId == config.ChannelId && config.Available) ||
			(request.ChannelId == "" && config.Method == request.Method && config.Available)

		if found {
			channelConfig = config
			logger.Log.Infof("find config: %v by request: %v", config, request)
			return
		}
	}
	err = fmt.Errorf("could'nt found available channel of appId: %v method: %v", request.AppId, request.Method)
	logger.Log.Errorf(err.Error())
	return
}

func (svc *PayGatewayService) RegisterGrpc(gs *grpc.Server) {
	pb.RegisterPayGatewayServer(gs, svc)
}

func Init(svc *service.Service) {
	payGatewayService := &PayGatewayService{}
	payGatewayService.Service = svc
	flag.Parse()

	gatewayConfig := service.InitGatewayConfig(svc.EtcdPeers, ETCD_DIR_ROOT)
	payGatewayService.GatewayConfig = gatewayConfig

	grpcClientFactory := gc.InitGrpFactory(*svc, payGatewayService.GatewayConfig)
	payGatewayService.GrpcClientFactory = grpcClientFactory

	payGatewayService.OrderGenerator = generator.New(gatewayConfig.PayConfig.ClusterId, gatewayConfig.PayConfig.Concurrency)
	payGatewayService.StartGrpc(payGatewayService.RegisterGrpc)
}
