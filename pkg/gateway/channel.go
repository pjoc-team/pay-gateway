package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
)

func (svc *PayGatewayService) GenerateChannelPayRequest(ctx context.Context, requestContext *RequestContext) (channelPayRequest *pay.ChannelPayRequest, err error) {
	log := logger.ContextLog(ctx)
	request := requestContext.PayRequest
	channelPayRequest = &pay.ChannelPayRequest{}
	if err = copier.Copy(channelPayRequest, request); err != nil {
		log.Errorf("Failed to copy struct from %v! error: %s", request, err.Error())
		return
	}
	channelPayRequest.GatewayOrderId = requestContext.GatewayOrderId

	if svc.payConfig.NotifyUrlPattern == "" {
		log.Errorf("NotifyUrlPattern is null!!!")
	}
	if svc.payConfig.ReturnUrlPattern == "" {
		log.Errorf("ReturnUrlPattern is null!!!")
	}
	// reset notify url
	channelPayRequest.NotifyUrl = ReplaceGatewayOrderId(svc.payConfig.NotifyUrlPattern, channelPayRequest.GatewayOrderId)
	channelPayRequest.ReturnUrl = ReplaceGatewayOrderId(svc.payConfig.ReturnUrlPattern, channelPayRequest.GatewayOrderId)
	channelPayRequest.ChannelAccount = requestContext.ChannelAccount
	channelPayRequest.PayAmount = request.GetPayAmount()
	product := &pay.Product{}
	product.Id = request.ProductId
	product.Name = request.ProductName
	product.Description = request.ProductDescribe
	channelPayRequest.Product = product
	channelPayRequest.UserIp = request.GetUserIp()
	channelPayRequest.Method = request.GetMethod()
	if extJson := request.ExtJson; extJson != "" {
		meta := make(map[string]string)
		if err = json.Unmarshal([]byte(extJson), &meta); err != nil {
			err = fmt.Errorf("failed to unmarshal json: %v error: %s", extJson, err.Error())
			log.Errorf(err.Error())
			return
		} else {
			channelPayRequest.Meta = meta
		}
	}
	return
}
