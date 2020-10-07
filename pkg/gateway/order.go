package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jinzhu/copier"
	"github.com/pjoc-team/pay-gateway/pkg/constant"
	"github.com/pjoc-team/pay-gateway/pkg/date"
	pb "github.com/pjoc-team/pay-proto/go"
	tracinglogger "github.com/pjoc-team/tracing/logger"
	"time"
)

func (svc *PayGatewayService) SavePayOrder(ctx context.Context, requestContext *RequestContext) (*pb.ReturnResult, error) {
	gatewayOrderId := requestContext.GatewayOrderId
	request := requestContext.PayRequest

	order := &pb.PayOrder{}
	requestContext.PayOrder = order
	basePayOrder := &pb.BasePayOrder{}
	order.BasePayOrder = basePayOrder
	copier.Copy(basePayOrder, request)

	order.OrderStatus = constant.ORDER_STATUS_WAITING
	basePayOrder.GatewayOrderId = gatewayOrderId
	basePayOrder.RequestTime = date.NowTime()
	basePayOrder.CreateDate = date.NowDate()
	basePayOrder.ChannelAccount = requestContext.ChannelAccount

	if serviceClient, e := svc.GetDatabaseClient(); e != nil {
		tracinglogger.ContextLog(ctx).Errorf("Failed to init database client! error: %s", e.Error())
		return nil, e
	} else {
		timeout, _ := context.WithTimeout(context.TODO(), 10*time.Second)
		if result, err := serviceClient.SavePayOrder(timeout, order); err != nil {
			tracinglogger.ContextLog(ctx).Errorf("Failed to save order: %v returns error: %s", order, err.Error())
			return nil, err
		} else {
			tracinglogger.ContextLog(ctx).Infof("Save db result: %v", result)
			return result, nil
		}
	}
}

func (svc *PayGatewayService) UpdatePayOrder(ctx context.Context, requestContext *RequestContext) (result *pb.ReturnResult, err error) {
	if requestContext.ChannelPayResponse == nil {
		tracinglogger.ContextLog(ctx).Errorf("Failed to update pay order! because channel response is null!")
		err = errors.New("failed update pay order!")
		return
	}
	order := requestContext.PayOrder
	if requestContext.ChannelPayResponse != nil {
		order.BasePayOrder.ChannelOrderId = requestContext.ChannelPayResponse.ChannelOrderId
	}

	if strings := requestContext.ChannelPayResponse.Data; strings != nil {
		if channelResponseJson, err := json.Marshal(strings); err != nil {
			tracinglogger.ContextLog(ctx).Errorf("Failed to marshal object: %v to json! error: %v", strings, err.Error())
		} else {
			order.BasePayOrder.ChannelResponseJson = string(channelResponseJson)
		}
	}

	svc.presentChannelErrorMessage(requestContext)

	if serviceClient, e := svc.GetDatabaseClient(); e != nil {
		tracinglogger.ContextLog(ctx).Errorf("Failed to init database client! error: %s", e.Error())
		return nil, e
	} else {
		timeout, _ := context.WithTimeout(context.TODO(), 10*time.Second)
		if result, err := serviceClient.UpdatePayOrder(timeout, order); err != nil {
			tracinglogger.ContextLog(ctx).Errorf("Failed to save order: %v returns error: %s", order, err.Error())
			return nil, err
		} else {
			tracinglogger.ContextLog(ctx).Infof("Save db result: %v", result)
			return result, nil
		}
	}
}

func (svc *PayGatewayService) presentChannelErrorMessage(requestContext *RequestContext) {
	if requestContext.err != nil {
		requestContext.PayOrder.BasePayOrder.ErrorMessage = requestContext.err.Error()
	}
}
