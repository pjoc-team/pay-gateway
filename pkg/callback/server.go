package callback

import (
	"context"
	"errors"
	"fmt"
	"github.com/pjoc-team/pay-gateway/pkg/discovery"
	pb "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
	"time"

	"net/http"
)

type NotifyService struct {
	services discovery.Services
}

func (svc *NotifyService) Notify(ctx context.Context, gatewayOrderId string,
	r *http.Request) (notifyResponse *pb.NotifyResponse, e error) {
	log := logger.ContextLog(ctx)

	var dbService pb.PayDatabaseServiceClient
	var putBackClientFunc discovery.PutBackClientFunc
	if dbService, putBackClientFunc, e = svc.services.GetDatabaseService(ctx); e != nil {
		log.Errorf("Failed to get db client! error: %v", e.Error())
		return
	}
	defer putBackClientFunc()
	orderQuery := &pb.PayOrder{BasePayOrder: &pb.BasePayOrder{GatewayOrderId: gatewayOrderId}}
	var existOrder *pb.PayOrder
	if response, err := dbService.FindPayOrder(ctx, orderQuery); err != nil {
		e = err
		log.Errorf("Failed to find order! error: %v order: %v", err.Error(), gatewayOrderId)
		return
	} else if response.PayOrders == nil || len(response.PayOrders) == 0 {
		log.Errorf("Not found order! order: %v", gatewayOrderId)
		e = fmt.Errorf("not found order: %v", gatewayOrderId)
		return
	} else {
		existOrder = response.PayOrders[0]
	}
	log.Infof("Processing order notify... order: %v", existOrder)
	// notify
	notifyResponse, e = svc.ProcessChannel(ctx, existOrder, r)
	if e != nil {
		return
	}

	settlementClient, e := svc.GetSettlementClient()
	if e != nil {
		log.Errorf("Failed to get settlement client! error: %v", e.Error())
		return
	} else if settlementClient == nil {
		log.Errorf("settlementClient is nil!")
		e = errors.New("system error")
		return
	}

	settlementRequest := &pb.SettlementPayOrder{Order: existOrder}
	timeoutSettle, _ := context.WithTimeout(context.TODO(), 10*time.Second)

	settlementResponse, e := settlementClient.ProcessOrderSuccess(timeoutSettle, settlementRequest)
	if e != nil {
		log.Errorf("Failed to settle order: %v error: %v", existOrder, e.Error())
		return
	} else {
		log.Infof("Notify order with result: %v", settlementResponse)
	}
	return
}

func (svc *NotifyService) ProcessChannel(ctx context.Context, existOrder *pb.PayOrder,
	r *http.Request) (notifyResponse *pb.NotifyResponse, e error) {
	log := logger.ContextLog(ctx)

	channelId := existOrder.BasePayOrder.ChannelId
	channelAccount := existOrder.BasePayOrder.ChannelAccount

	// send to channel client
	var client pb.PayChannelClient
	var pubBackClientFunc discovery.PutBackClientFunc
	if client, pubBackClientFunc, e = svc.services.GetChannelClient(ctx, channelId); e != nil {
		log.Errorf("Failed to get channel client of channelId: %v! error: %v", channelId, e.Error())
		return
	}
	if pubBackClientFunc != nil {
		defer pubBackClientFunc()
	}
	var request *pb.HTTPRequest
	if request, e = BuildChannelHttpRequest(ctx, r); e != nil {
		log.Errorf("Failed to build notify request! error: %v", e.Error())
		return
	}
	notifyRequest := &pb.NotifyRequest{PaymentAccount: channelAccount, Request: request, Type: pb.PayType_PAY, Method: existOrder.BasePayOrder.Method}

	timeoutChannel, _ := context.WithTimeout(context.TODO(), 10*time.Second)
	if notifyResponse, e = client.Notify(timeoutChannel, notifyRequest); e != nil {
		log.Errorf("Failed to notify channel! order: %v error: %v", existOrder, e.Error())
		return
	} else {
		log.Infof("Notify to channel: %v with result: %v", channelId, notifyResponse)
	}
	return
}

func Init(services discovery.Services) *NotifyService {
	notify := &NotifyService{}
	notify.services = services

	return notify
}
