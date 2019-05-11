package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jinzhu/copier"
	"github.com/pjoc-team/base-service/pkg/constant"
	"github.com/pjoc-team/base-service/pkg/date"
	"github.com/pjoc-team/base-service/pkg/logger"
	pb "github.com/pjoc-team/pay-proto/go"
	"time"
)

func (svc *PayGatewayService) SavePayOrder(requestContext *RequestContext) (*pb.ReturnResult, error) {
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
		logger.Log.Errorf("Failed to init database client! error: %s", e.Error())
		return nil, e
	} else {
		timeout, _ := context.WithTimeout(context.TODO(), 10*time.Second)
		if result, err := serviceClient.SavePayOrder(timeout, order); err != nil {
			logger.Log.Errorf("Failed to save order: %v returns error: %s", order, err.Error())
			return nil, err
		} else {
			logger.Log.Infof("Save db result: %v", result)
			return result, nil
		}
	}
}

func (svc *PayGatewayService) UpdatePayOrder(requestContext *RequestContext) (result *pb.ReturnResult, err error) {
	if requestContext.ChannelPayResponse == nil {
		logger.Log.Errorf("Failed to update pay order! because channel response is null!")
		err = errors.New("failed update pay order!")
		return
	}
	order := requestContext.PayOrder
	if requestContext.ChannelPayResponse != nil {
		order.BasePayOrder.ChannelOrderId = requestContext.ChannelPayResponse.ChannelOrderId
	}

	if strings := requestContext.ChannelPayResponse.Data; strings != nil {
		if channelResponseJson, err := json.Marshal(strings); err != nil {
			logger.Log.Errorf("Failed to marshal object: %v to json! error: %v", strings, err.Error())
		} else {
			order.BasePayOrder.ChannelResponseJson = string(channelResponseJson)
		}
	}

	svc.presentChannelErrorMessage(requestContext)

	if serviceClient, e := svc.GetDatabaseClient(); e != nil {
		logger.Log.Errorf("Failed to init database client! error: %s", e.Error())
		return nil, e
	} else {
		timeout, _ := context.WithTimeout(context.TODO(), 10*time.Second)
		if result, err := serviceClient.UpdatePayOrder(timeout, order); err != nil {
			logger.Log.Errorf("Failed to save order: %v returns error: %s", order, err.Error())
			return nil, err
		} else {
			logger.Log.Infof("Save db result: %v", result)
			return result, nil
		}
	}
}

func (svc *PayGatewayService) presentChannelErrorMessage(requestContext *RequestContext) {
	if requestContext.err != nil {
		requestContext.PayOrder.BasePayOrder.ErrorMessage = requestContext.err.Error()
	}
}
