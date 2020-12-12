package callback

import (
	"context"
	"fmt"
	"github.com/pjoc-team/pay-gateway/pkg/discovery"
	"github.com/pjoc-team/pay-gateway/pkg/metadata"
	md "google.golang.org/grpc/metadata"

	pb "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
	"github.com/pjoc-team/tracing/tracing"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"net/http"
)

// NotifyService notify service
type NotifyService struct {
	services *discovery.Services
}

// CallbackByGet callback by get
func (svc *NotifyService) CallbackByGet(
	request *pb.HttpCallbackRequest, stream pb.ChannelCallback_CallbackByGetServer,
) error {
	log := logger.Log()
	log.Infof("request: %v", request)
	headers, ok := metadata.GrpcGatewayHeaders(stream.Context())
	// head, ok := metadata.FromIncomingContext(stream.Context())
	if ok {
		log.Infof("head: %v", headers)
	}
	resp := &httpbody.HttpBody{
		ContentType: "text/html",
		Data:        []byte("get ok"),
	}
	respHead := md.New(
		map[string]string{
			"count": fmt.Sprintf("%d", len(resp.Data)),
		},
	)
	err := stream.SendHeader(respHead)
	if err != nil {
		log.Errorf("failed to send")
		return nil
	}

	err = stream.Send(resp)
	if err != nil {
		log.Errorf("failed to send: %v error: %v", resp, err.Error())
	}
	return nil
}

// CallbackByPost callback by posts
func (svc *NotifyService) CallbackByPost(
	request *pb.HttpCallbackRequest, stream pb.ChannelCallback_CallbackByPostServer,
) error {
	log := logger.Log()
	headers, ok := metadata.GrpcGatewayHeaders(stream.Context())
	// head, ok := metadata.FromIncomingContext(stream.Context())
	if ok {
		log.Infof("head: %v", headers)
	}

	log.Infof("request: %v", request)
	resp := &httpbody.HttpBody{
		ContentType: "text/html",
		Data:        []byte("post ok"),
	}
	respHead := md.New(
		map[string]string{
			"count": fmt.Sprintf("%d", len(resp.Data)),
		},
	)
	err := stream.SendHeader(respHead)
	if err != nil {
		log.Errorf("failed to send")
		return nil
	}
	err = stream.Send(resp)
	if err != nil {
		log.Errorf("failed to send")
		return nil
	}
	return nil
}

// Notify notify by order id
func (svc *NotifyService) Notify(
	ctx context.Context, gatewayOrderID string,
	r *http.Request,
) (notifyResponse *pb.ChannelNotifyResponse, e error) {
	span, ctx := tracing.Start(ctx, "notify")
	defer span.Finish()

	log := logger.ContextLog(ctx)

	var dbService pb.PayDatabaseServiceClient
	var putBackClientFunc discovery.PutBackClientFunc
	if dbService, putBackClientFunc, e = svc.services.GetDatabaseService(ctx); e != nil {
		log.Errorf("Failed to get db client! error: %v", e.Error())
		return
	}
	defer putBackClientFunc()
	orderQuery := &pb.PayOrder{BasePayOrder: &pb.BasePayOrder{GatewayOrderId: gatewayOrderID}}
	var existOrder *pb.PayOrder
	if response, err := dbService.FindPayOrder(ctx, orderQuery); err != nil {
		e = err
		log.Errorf("Failed to find order! error: %v order: %v", err.Error(), gatewayOrderID)
		return
	} else if response.PayOrders == nil || len(response.PayOrders) == 0 {
		log.Errorf("Not found order! order: %v", gatewayOrderID)
		e = fmt.Errorf("not found order: %v", gatewayOrderID)
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

	settlementClient, pubBackClientFunc, e := svc.services.GetSettlementClient(ctx)
	if e != nil {
		log.Errorf("Failed to get settlement client! error: %v", e.Error())
		return
	}
	defer pubBackClientFunc()

	settlementRequest := &pb.SettlementPayOrder{Order: existOrder}
	settlementResponse, e := settlementClient.ProcessOrderSuccess(ctx, settlementRequest)
	if e != nil {
		log.Errorf("Failed to settle order: %v error: %v", existOrder, e.Error())
		return
	}
	log.Infof("Notify order with result: %v", settlementResponse)
	return
}

// ProcessChannel process channel's callback
func (svc *NotifyService) ProcessChannel(
	ctx context.Context, existOrder *pb.PayOrder,
	r *http.Request,
) (notifyResponse *pb.ChannelNotifyResponse, e error) {
	log := logger.ContextLog(ctx)

	channelID := existOrder.BasePayOrder.ChannelId
	channelAccount := existOrder.BasePayOrder.ChannelAccount

	// send to channel client
	var client pb.PayChannelClient
	var pubBackClientFunc discovery.PutBackClientFunc
	if client, pubBackClientFunc, e = svc.services.GetChannelClient(ctx, channelID); e != nil {
		log.Errorf("Failed to get channel client of channelID: %v! error: %v", channelID, e.Error())
		return
	}
	if pubBackClientFunc != nil {
		defer pubBackClientFunc()
	}
	var request *pb.HTTPRequest
	if request, e = BuildChannelHTTPRequest(ctx, r); e != nil {
		log.Errorf("Failed to build notify request! error: %v", e.Error())
		return
	}
	notifyRequest := &pb.ChannelNotifyRequest{
		PaymentAccount: channelAccount, Request: request, Type: pb.PayType_PAY,
		Method: existOrder.BasePayOrder.Method,
	}

	if notifyResponse, e = client.ChannelNotify(ctx, notifyRequest); e != nil {
		log.Errorf("Failed to notify channel! order: %v error: %v", existOrder, e.Error())
		return
	}
	log.Infof("Notify to channel: %v with result: %v", channelID, notifyResponse)
	return
}

// NewServer init notify service
func NewServer(services *discovery.Services) (pb.ChannelCallbackServer, error) {
	notify := &NotifyService{}
	notify.services = services

	return notify, nil
}
