package settlement

import (
	"errors"
	"fmt"
	"github.com/pjoc-team/pay-gateway/pkg/configclient"
	"github.com/pjoc-team/pay-gateway/pkg/constant"
	"github.com/pjoc-team/pay-gateway/pkg/discovery"
	"github.com/pjoc-team/pay-gateway/pkg/generator"
	"github.com/pjoc-team/pay-gateway/pkg/notify"
	pb "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
	"golang.org/x/net/context"
	"time"
)

type service struct {
	config         configclient.ConfigClients
	OrderGenerator *generator.Generator
	services       *discovery.Services
	notifyService  *notify.Service
	scheduler      *notify.Scheduler
}

func (svc *service) SendNotify(ctx context.Context, notify *pb.PayNotice) (
	result *pb.ReturnResult, err error,
) {
	log := logger.ContextLog(ctx)
	if notify == nil {
		err = errors.New("notify is nil")
		return
	}
	err = svc.notifyService.Notify(ctx, notify)
	if err != nil {
		log.Errorf("Failed to send notify! error: %v", err.Error())
		return
	}

	return
}

func (svc *service) NotifyOrder(
	ctx context.Context, settlementPayOrder *pb.SettlementPayOrder,
) (response *pb.SettlementResponse, err error) {
	log := logger.ContextLog(ctx)
	dbClient, pubBackClientFunc, err := svc.services.GetDatabaseService(ctx)
	if err != nil {
		log.Errorf("Failed to get db client! error: %v", err.Error())
		return
	}
	defer pubBackClientFunc()

	timeoutCtx, _ := context.WithTimeout(ctx, 6*time.Second)
	order := &pb.PayOrder{}
	order.BasePayOrder = &pb.BasePayOrder{GatewayOrderId: settlementPayOrder.Order.BasePayOrder.GatewayOrderId}

	orderResponse, err := dbClient.FindPayOrder(timeoutCtx, order)
	if err != nil {
		log.Errorf(
			"Failed to find order: %v", settlementPayOrder.Order.BasePayOrder.GatewayOrderId,
		)
		return
	}
	if orderResponse.PayOrders == nil || len(orderResponse.PayOrders) == 0 {
		log.Errorf("Not found order! request: %v", settlementPayOrder)
		err = fmt.Errorf("failed to found order of: %v", settlementPayOrder)
		return
	}
	existsOrder := orderResponse.PayOrders[0]
	if existsOrder.OrderStatus != constant.OrderStatusSuccess {
		err = fmt.Errorf("order: %v is not success", existsOrder.BasePayOrder.GatewayOrderId)
		log.Error(err.Error())
		return
	}
	payNotice := svc.notifyService.GeneratePayNotice(existsOrder)

	timeoutCtxNotice, _ := context.WithTimeout(ctx, 6*time.Second)
	result, err := dbClient.SavePayNotice(timeoutCtxNotice, payNotice)
	if err != nil {
		log.Errorf("Failed to save payNotice: %v", payNotice)
		return
	} else if result != nil && result.Code != pb.ReturnResultCode_CODE_SUCCESS {
		err = fmt.Errorf("failed to save payNotice! message: %v", result)
		log.Errorf(err.Error())
		return
	}
	returnResult, err := svc.SendNotify(ctx, payNotice)
	if err != nil {
		log.Errorf("Failed to save payNotice: %v", payNotice)
		return
	} else if returnResult != nil && returnResult.Code != pb.ReturnResultCode_CODE_SUCCESS {
		err = fmt.Errorf("failed to save payNotice! message: %v", result)
		log.Errorf(err.Error())
		return
	}
	response = &pb.SettlementResponse{}
	response.Result = returnResult
	return
}

func (svc *service) ProcessOrderSuccess(
	ctx context.Context, settlementPayOrder *pb.SettlementPayOrder,
) (response *pb.SettlementResponse, err error) {
	log := logger.ContextLog(ctx)
	response = &pb.SettlementResponse{}
	dbClient, pubBackClientFunc, err := svc.services.GetDatabaseService(ctx)
	if err != nil {
		log.Errorf("Failed to get db client! error: %v", err.Error())
		return
	}
	defer pubBackClientFunc()
	orderResponse, err := dbClient.FindPayOrder(ctx, settlementPayOrder.Order)
	if err != nil {
		log.Errorf(
			"Failed to find order: %v", settlementPayOrder.Order.BasePayOrder.GatewayOrderId,
		)
		return
	}
	existsOrder := orderResponse.PayOrders[0]
	if existsOrder.OrderStatus == constant.OrderStatusSuccess {
		response = &pb.SettlementResponse{}
		response.Result = &pb.ReturnResult{Code: pb.ReturnResultCode_CODE_SUCCESS}
		return response, nil
	}

	sendStatus := constant.NotifyFailed
	// save to order ok
	orderOk, _ := svc.GenerateSuccessOrder(ctx, existsOrder)
	orderOk.SendNoticeStats = sendStatus

	result, e := dbClient.SavePayOrderOk(ctx, orderOk)
	if e != nil {
		log.Errorf("Failed to save order ok!")
		return
	} else if result != nil && result.Code != pb.ReturnResultCode_CODE_SUCCESS {
		err = fmt.Errorf("failed to save order ok! message: %v", result)
		log.Errorf(err.Error())
		return
	}

	// notify
	settlementResponse, err := svc.NotifyOrder(ctx, settlementPayOrder)
	if err != nil {
		log.Errorf(
			"Failed to notify order! error: %v order: %v", err.Error(), settlementPayOrder,
		)
		return nil, err
	} else if settlementResponse.Result != nil && settlementResponse.Result.Code != pb.ReturnResultCode_CODE_SUCCESS {
		err = fmt.Errorf("failed to notify order! message: %v", settlementResponse.Result)
		log.Errorf(err.Error())
		return
	}
	sendStatus = constant.NotifySuccess
	orderOk.SendNoticeStats = sendStatus
	result, err = dbClient.UpdatePayOrderOk(ctx, orderOk)
	if err != nil {
		log.Errorf("Failed to update order ok! order: %v error: %v", orderOk, err.Error())
		response.Result = result
		return
	}

	response.Result = &pb.ReturnResult{Code: pb.ReturnResultCode_CODE_SUCCESS}
	return response, nil
}

// ProcessSuccess process success order
func (svc *service) ProcessSuccess(
	ctx context.Context, request *pb.SettlementRequest,
) (*pb.SettlementResponse, error) {
	gatewayOrderID := request.GatewayOrderId
	order := &pb.PayOrder{BasePayOrder: &pb.BasePayOrder{GatewayOrderId: gatewayOrderID}}
	settlementPayOrder := &pb.SettlementPayOrder{Order: order}
	return svc.ProcessOrderSuccess(ctx, settlementPayOrder)
}

// New create settlement gateway
func New(
	services *discovery.Services, scheduler *notify.Scheduler, cc configclient.ConfigClients,
	notifyService *notify.Service,
) (pb.SettlementGatewayServer, error) {
	s := &service{}
	s.services = services
	s.notifyService = notifyService
	s.config = cc
	s.scheduler = scheduler
	return s, nil
}
