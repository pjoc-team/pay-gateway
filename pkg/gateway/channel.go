package gateway

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/pjoc-team/base-service/pkg/logger"
	"github.com/pjoc-team/pay-proto/go"
)

func (svc *PayGatewayService) GenerateChannelPayRequest(ctx *RequestContext) (channelPayRequest *pay.ChannelPayRequest, err error) {
	request := ctx.PayRequest
	channelPayRequest = &pay.ChannelPayRequest{}
	if err = copier.Copy(channelPayRequest, request); err != nil {
		logger.Log.Errorf("Failed to copy struct from %v! error: %s", request, err.Error())
		return
	}
	channelPayRequest.GatewayOrderId = ctx.GatewayOrderId

	if svc.PayConfig.NotifyUrlPattern == "" {
		logger.Log.Errorf("NotifyUrlPattern is null!!!")
	}
	if svc.PayConfig.ReturnUrlPattern == "" {
		logger.Log.Errorf("ReturnUrlPattern is null!!!")
	}
	// reset notify url
	channelPayRequest.NotifyUrl = ReplaceGatewayOrderId(svc.PayConfig.NotifyUrlPattern, channelPayRequest.GatewayOrderId)
	channelPayRequest.ReturnUrl = ReplaceGatewayOrderId(svc.PayConfig.ReturnUrlPattern, channelPayRequest.GatewayOrderId)
	channelPayRequest.ChannelAccount = ctx.ChannelAccount
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
			logger.Log.Errorf(err.Error())
			return
		} else {
			channelPayRequest.Meta = meta
		}
	}
	return
}
