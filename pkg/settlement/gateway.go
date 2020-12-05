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

func (svc *service) SendNotice(ctx context.Context, notice *pb.PayNotice) (
	result *pb.ReturnResult, err error,
) {
	log := logger.ContextLog(ctx)
	if notice == nil {
		err = errors.New("notice is nil")
		return
	}
	err = svc.notifyService.Notice(notice)
	if err != nil {
		log.Errorf("Failed to send notice! error: %v", err.Error())
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
	returnResult, err := svc.SendNotice(ctx, payNotice)
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
	dbClient, pubBackClientFunc, err := svc.services.GetDatabaseService(ctx)
	if err != nil {
		log.Errorf("Failed to get db client! error: %v", err.Error())
		return
	}
	defer pubBackClientFunc()
	timeoutCtx, _ := context.WithTimeout(ctx, 6*time.Second)
	orderResponse, err := dbClient.FindPayOrder(timeoutCtx, settlementPayOrder.Order)
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
	timeoutSaveCtx, _ := context.WithTimeout(ctx, 6*time.Second)
	orderOk := svc.GenerateSuccessOrder(existsOrder)
	orderOk.SendNoticeStats = sendStatus

	result, e := dbClient.SavePayOrderOk(timeoutSaveCtx, orderOk)
	if e != nil {
		log.Errorf("Failed to save order ok!")
		return
	} else if result != nil && result.Code != pb.ReturnResultCode_CODE_SUCCESS {
		err = fmt.Errorf("failed to save order ok! message: %v", result)
		log.Errorf(err.Error())
		return
	}

	// notify
	timeoutNotifyCtx, _ := context.WithTimeout(ctx, 6*time.Second)
	settlementResponse, err := svc.NotifyOrder(timeoutNotifyCtx, settlementPayOrder)
	if err != nil {
		log.Errorf(
			"Failed to notify order! error: %v order: %v", err.Error(), settlementPayOrder,
		)
		sendStatus = constant.NotifyFailed
		return
	} else if settlementResponse.Result != nil && settlementResponse.Result.Code != pb.ReturnResultCode_CODE_SUCCESS {
		err = fmt.Errorf("failed to notify order! message: %v", settlementResponse.Result)
		log.Errorf(err.Error())
		return
	}
	sendStatus = constant.NotifySuccess
	orderOk.SendNoticeStats = sendStatus
	timeoutUpdateCtx, _ := context.WithTimeout(ctx, 6*time.Second)
	result, err = dbClient.UpdatePayOrderOk(timeoutUpdateCtx, orderOk)
	if err != nil {
		log.Errorf("Failed to update order ok! order: %v error: %v", orderOk, err.Error())
		return
	}

	response = &pb.SettlementResponse{}
	response.Result = &pb.ReturnResult{Code: pb.ReturnResultCode_CODE_SUCCESS}
	return response, nil
}

func (svc *service) ProcessSuccess(
	ctx context.Context, request *pb.SettlementRequest,
) (*pb.SettlementResponse, error) {
	gatewayOrderId := request.GatewayOrderId
	order := &pb.PayOrder{BasePayOrder: &pb.BasePayOrder{GatewayOrderId: gatewayOrderId}}
	settlementPayOrder := &pb.SettlementPayOrder{Order: order}
	return svc.ProcessOrderSuccess(ctx, settlementPayOrder)
}

func New(
	services *discovery.Services, scheduler *notify.Scheduler, cc configclient.ConfigClients,
	notifyService *notify.Service,
) *service {
	s := &service{}
	s.services = services
	s.notifyService = notifyService
	s.config = cc
	s.scheduler = scheduler
	return s
}
