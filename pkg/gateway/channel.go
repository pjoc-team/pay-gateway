package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/blademainer/commons/pkg/util"
	"github.com/jinzhu/copier"
	"github.com/pjoc-team/pay-gateway/pkg/configclient"
	"github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
)

// GenerateChannelPayRequest generate pay request
func (svc *PayGatewayService) GenerateChannelPayRequest(
	ctx context.Context, requestContext *RequestContext,
) (channelPayRequest *pay.ChannelPayRequest, err error) {
	log := logger.ContextLog(ctx)
	request := requestContext.PayRequest
	channelPayRequest = &pay.ChannelPayRequest{}
	if err = copier.Copy(channelPayRequest, request); err != nil {
		log.Errorf("Failed to copy struct from %v! error: %s", request, err.Error())
		return
	}
	channelPayRequest.GatewayOrderId = requestContext.GatewayOrderID

	payConfig, err := svc.configClients.GetPayConfig(ctx)
	if err != nil {
		log.Errorf("failed to pay config, error: %v", err.Error())
		return nil, err
	}
	// reset notify url
	channelPayRequest.NotifyUrl, err = svc.notifyURL(ctx, requestContext, payConfig)
	if err != nil {
		log.Errorf("failed to get notify url, error: %v", err.Error())
		return nil, err
	}
	channelPayRequest.ReturnUrl = ReplaceGatewayOrderID(payConfig.ReturnURLPattern, channelPayRequest.GatewayOrderId)
	channelPayRequest.ChannelAccount = requestContext.ChannelAccount
	channelPayRequest.PayAmount = request.GetPayAmount()
	product := &pay.Product{}
	product.Id = request.ProductId
	product.Name = request.ProductName
	product.Description = request.ProductDescribe
	channelPayRequest.Product = product
	channelPayRequest.UserIp = request.GetUserIp()
	channelPayRequest.Method = request.GetMethod()
	if extJSON := request.ExtJson; extJSON != "" {
		meta := make(map[string]string)
		if err = json.Unmarshal([]byte(extJSON), &meta); err != nil {
			err = fmt.Errorf("failed to unmarshal json: %v error: %s", extJSON, err.Error())
			log.Errorf(err.Error())
			return
		}
		channelPayRequest.Meta = meta
	}
	return
}

func (svc *PayGatewayService) notifyURL(
	ctx context.Context, requestContext *RequestContext, config *configclient.PayConfig,
) (string, error) {
	kv := map[string]string{
		"order_id": requestContext.GatewayOrderID,
		"account":  requestContext.ChannelAccount,
		"channel":  requestContext.PayRequest.GetChannelId(),
	}
	notifyURL, err := util.ReplacePlaceholders(config.NotifyURLPattern, kv)
	if err != nil {
		return "", err
	}
	return notifyURL, nil
}
