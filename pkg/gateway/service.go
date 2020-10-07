package gateway

import (
	"flag"
	"fmt"
	"github.com/pjoc-team/pay-gateway/pkg/config"
	"github.com/pjoc-team/pay-gateway/pkg/generator"
	"github.com/pjoc-team/pay-gateway/pkg/validator"
	pb "github.com/pjoc-team/pay-proto/go"
	tracinglogger "github.com/pjoc-team/tracing/logger"
	"golang.org/x/net/context"
)

type PayGatewayService struct {
	OrderGenerator *generator.Generator
	config         config.Server
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
	tracinglogger.ContextLog(ctx).Debugf("New request: %v", request)
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

	if result, e := svc.SavePayOrder(ctx, requestContext); e != nil || result.Code != pb.ReturnResultCode_CODE_SUCCESS {
		err = e
		response.Result = result
		return BuildSystemErrorResponse(err), nil
	}

	var client pb.PayChannelClient
	client, err = svc.GetChannelClient(request.GetChannelId())
	if client == nil || err != nil {
		tracinglogger.ContextLog(ctx).Errorf("Failed to get channelClient! channelId: %s, error: %s, ", request.GetChannelId(), err.Error())
		return BuildSystemErrorResponse(err), nil
	} else {
		tracinglogger.ContextLog(ctx).Debugf("Got client: %v for channelId: %s", client, request.GetChannelId())
	}
	var channelPayRequest *pb.ChannelPayRequest
	if channelPayRequest, err = svc.GenerateChannelPayRequest(requestContext); err != nil {
		return BuildSystemErrorResponse(err), nil
	}
	var channelPayResponse *pb.ChannelPayResponse
	if channelPayResponse, err = client.Pay(ctx, channelPayRequest); err != nil {
		tracinglogger.ContextLog(ctx).Errorf("Pay channel failed! err: %s channelPayResponse: %v", err.Error(), channelPayResponse)
		requestContext.err = err
		svc.UpdatePayOrder(ctx, requestContext)
		return BuildSystemErrorResponse(err), nil
	} else if channelPayResponse == nil || channelPayResponse.Data == nil {
		err = fmt.Errorf("channel response fail! response: %v", channelPayResponse)
		tracinglogger.ContextLog(ctx).Errorf(err.Error())
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
func (svc *PayGatewayService) processChannelIdIfNotPresent(ctx context.Context, request *pb.PayRequest) (channelConfig AppIdChannelConfig, err error) {
	log := tracinglogger.ContextLog(ctx)

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
			log.Infof("find config: %v by request: %v", config, request)
			return
		}
	}
	err = fmt.Errorf("could'nt found available channel of appId: %v method: %v", request.AppId, request.Method)
	log.Errorf(err.Error())
	return
}

func NewPayGateway(config config.Server, clusterID string, concurrency int) (pb.PayGateway, error) {
	flag.Parse()
	payGatewayService := &PayGatewayService{}
	payGatewayService.config = config
	payGatewayService.OrderGenerator = generator.New(clusterID, concurrency)
	return payGatewayService, nil
}
