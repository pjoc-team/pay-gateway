package gateway

import (
	"fmt"
	"github.com/pjoc-team/pay-gateway/pkg/configclient"
	"github.com/pjoc-team/pay-gateway/pkg/discovery"
	"github.com/pjoc-team/pay-gateway/pkg/generator"
	"github.com/pjoc-team/pay-gateway/pkg/validator"
	pb "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
	"golang.org/x/net/context"
)

// PayGatewayService pay gateway service
type PayGatewayService struct {
	configClients  configclient.ConfigClients
	orderGenerator *generator.Generator
	services       *discovery.Services
}

// RequestContext context of request
type RequestContext struct {
	GatewayOrderID     string
	ChannelAccount     string
	PayRequest         *pb.PayRequest
	PayOrder           *pb.PayOrder
	ChannelPayRequest  *pb.ChannelPayRequest
	ChannelPayResponse *pb.ChannelPayResponse
	err                error
}

// BuildParamsErrorResponse build params error
func BuildParamsErrorResponse(err error) *pb.PayResponse {
	response := &pb.PayResponse{}
	response.Result = &pb.ReturnResult{
		Code: pb.ReturnResultCode_CODE_PARAMS_ERROR, Message: "PARAMS_ERROR", Describe: err.Error(),
	}
	return response
}

// BuildSystemErrorResponse builder system error
func BuildSystemErrorResponse(err error) *pb.PayResponse {
	response := &pb.PayResponse{}
	response.Result = &pb.ReturnResult{
		Code: pb.ReturnResultCode_CODE_SYSTEM_ERROR, Message: "SYSTEM_ERROR", Describe: err.Error(),
	}
	return response
}

// Pay process pay request
func (svc *PayGatewayService) Pay(
	ctx context.Context, request *pb.PayRequest,
) (response *pb.PayResponse, err error) {
	log := logger.ContextLog(ctx)
	log.Debugf("new request: %v", request)
	if err = request.Validate(); err != nil {
		return BuildParamsErrorResponse(err), nil
	}
	if err = validator.Validate(ctx, *request, svc.configClients.GetAppConfig); err != nil {
		return BuildParamsErrorResponse(err), nil
	}
	var cfg *configclient.AppIDChannelConfig
	if cfg, err = svc.processChannelIDIfNotPresent(ctx, request); err != nil {
		log.Error(err.Error())
		err = fmt.Errorf("could'nt found config of channelID: %v", request.ChannelId)
		return BuildParamsErrorResponse(err), nil
	}

	response = &pb.PayResponse{}
	requestContext := &RequestContext{}

	gatewayOrderID := svc.orderGenerator.GenerateID()
	requestContext.GatewayOrderID = gatewayOrderID
	requestContext.PayRequest = request
	requestContext.ChannelAccount = cfg.ChannelAccount

	if result, e := svc.SavePayOrder(
		ctx, requestContext,
	); e != nil || result.Code != pb.ReturnResultCode_CODE_SUCCESS {
		err = e
		response.Result = result
		return BuildSystemErrorResponse(err), nil
	}

	var client pb.PayChannelClient
	client, err = svc.services.GetChannelClient(ctx, request.GetChannelId())
	if client == nil || err != nil {
		log.Errorf(
			"Failed to get channelClient! channelID: %s, error: %s, ", request.GetChannelId(),
			err.Error(),
		)
		return BuildSystemErrorResponse(err), nil
	}
	log.Debugf("Got client: %v for channelID: %s", client, request.GetChannelId())
	var channelPayRequest *pb.ChannelPayRequest
	if channelPayRequest, err = svc.GenerateChannelPayRequest(ctx, requestContext); err != nil {
		return BuildSystemErrorResponse(err), nil
	}
	var channelPayResponse *pb.ChannelPayResponse
	if channelPayResponse, err = client.Pay(ctx, channelPayRequest); err != nil {
		log.Errorf(
			"Pay channel failed! err: %s channelPayResponse: %v", err.Error(), channelPayResponse,
		)
		requestContext.err = err
		order, err := svc.UpdatePayOrder(ctx, requestContext)
		if err != nil {
			log.Errorf("failed to update pay order: %#v error: %v", order, err.Error())
		}
		return BuildSystemErrorResponse(err), nil
	} else if channelPayResponse == nil {
		err = fmt.Errorf("channel response fail! response: %v", channelPayResponse)
		log.Errorf(err.Error())
		return BuildSystemErrorResponse(err), nil
	}
	requestContext.ChannelPayResponse = channelPayResponse
	if result, e := svc.UpdatePayOrder(
		ctx, requestContext,
	); e != nil || result.Code != pb.ReturnResultCode_CODE_SUCCESS {
		response.Result = result
		err = e
		return
	}

	response.Data = channelPayResponse.Data
	response.GatewayOrderId = gatewayOrderID
	response.Result = SuccessResult

	return response, nil
}

// 如果没有传入channelID，则根据method找可用的channelID
func (svc *PayGatewayService) processChannelIDIfNotPresent(
	ctx context.Context, request *pb.PayRequest,
) (channelConfig *configclient.AppIDChannelConfig, err error) {
	log := logger.ContextLog(ctx)

	channelConfigs, err := svc.configClients.GetAppChannelConfig(
		ctx, request.AppId, request.Method.String(),
	)
	if err != nil {
		err = fmt.Errorf("failed to found info of appID: %v", request.AppId)
		return
	} else if len(channelConfigs) == 0 {
		err = fmt.Errorf(
			"failed to get config of appID: %v method: %v", request.AppId, request.Method.String(),
		)
		log.Error(err.Error())
		return
	}

	for _, config := range channelConfigs {
		found := (request.ChannelId != "" && request.ChannelId == config.ChannelID && config.Available) ||
			(request.ChannelId == "" && config.Method == request.Method.String() && config.Available)
		if found {
			channelConfig = config
			request.ChannelId = config.ChannelID
			log.Infof("find config: %v by request: %v", config, request)
			return
		}
	}
	err = fmt.Errorf(
		"failed to get config of appID: %v method: %v", request.AppId, request.Method.String(),
	)
	return
}

// NewPayGateway new gateway service
func NewPayGateway(
	cc configclient.ConfigClients, clusterID string, concurrency int, services *discovery.Services,
) (pb.PayGatewayServer, error) {
	payGatewayService := &PayGatewayService{}
	payGatewayService.configClients = cc
	payGatewayService.orderGenerator = generator.New(clusterID, concurrency)
	payGatewayService.services = services
	return payGatewayService, nil
}
