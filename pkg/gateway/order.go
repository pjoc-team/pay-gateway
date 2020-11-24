package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jinzhu/copier"
	"github.com/pjoc-team/pay-gateway/pkg/constant"
	"github.com/pjoc-team/pay-gateway/pkg/date"
	"github.com/pjoc-team/pay-gateway/pkg/discovery"
	pb "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
)

// SavePayOrder save pay order
func (svc *PayGatewayService) SavePayOrder(ctx context.Context, requestContext *RequestContext) (*pb.ReturnResult, error) {
	log := logger.ContextLog(ctx)

	var putBackClientFunc discovery.PutBackClientFunc
	dbService, putBackClientFunc, err2 := svc.services.GetDatabaseService(ctx)
	if err2 != nil {
		log.Errorf("failed to get db service, err: %v", err2.Error())
		return nil, err2
	}
	defer putBackClientFunc()

	gatewayOrderID := requestContext.GatewayOrderID
	request := requestContext.PayRequest

	order := &pb.PayOrder{}
	requestContext.PayOrder = order
	basePayOrder := &pb.BasePayOrder{}
	order.BasePayOrder = basePayOrder
	err := copier.Copy(basePayOrder, request)
	if err != nil {
		return nil, err
	}

	order.OrderStatus = constant.OrderStatusWaiting
	basePayOrder.GatewayOrderId = gatewayOrderID
	basePayOrder.RequestTime = date.NowTime()
	basePayOrder.CreateDate = date.NowDate()
	basePayOrder.ChannelAccount = requestContext.ChannelAccount

	result, err := dbService.SavePayOrder(ctx, order)
	if err != nil {
		log.Errorf("failed to save order: %v returns error: %s", order, err.Error())
		return nil, err
	}
	log.Infof("save db result: %v", result)
	return result, nil
}

// UpdatePayOrder update pay order
func (svc *PayGatewayService) UpdatePayOrder(ctx context.Context, requestContext *RequestContext) (result *pb.ReturnResult, err error) {
	log := logger.ContextLog(ctx)
	dbService, putBackClientFunc, err2 := svc.services.GetDatabaseService(ctx)
	if err2 != nil {
		log.Errorf("failed to get db service, err: %v", err2.Error())
		return nil, err2
	}
	defer putBackClientFunc()

	if requestContext.ChannelPayResponse == nil {
		log.Errorf("failed to update pay order! because channel response is null!")
		err = errors.New("failed update pay order")
		return
	}
	order := requestContext.PayOrder
	if requestContext.ChannelPayResponse != nil {
		order.BasePayOrder.ChannelOrderId = requestContext.ChannelPayResponse.ChannelOrderId
	}

	if strings := requestContext.ChannelPayResponse.Data; strings != nil {
		if channelResponseJSON, err := json.Marshal(strings); err != nil {
			log.Errorf("Failed to marshal object: %v to json! error: %v", strings, err.Error())
		} else {
			order.BasePayOrder.ChannelResponseJson = string(channelResponseJSON)
		}
	}

	svc.presentChannelErrorMessage(requestContext)

	result, err = dbService.UpdatePayOrder(ctx, order)
	if err != nil {
		log.Errorf("Failed to save order: %v returns error: %s", order, err.Error())
		return nil, err
	}
	log.Infof("Save db result: %v", result)
	return result, nil
}

func (svc *PayGatewayService) presentChannelErrorMessage(requestContext *RequestContext) {
	if requestContext.err != nil {
		requestContext.PayOrder.BasePayOrder.ErrorMessage = requestContext.err.Error()
	}
}
