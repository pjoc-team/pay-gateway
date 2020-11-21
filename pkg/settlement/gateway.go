package settlement

import (
	"errors"
	"flag"
	"fmt"
	"gitlab.com/pjoc/base-service/pkg/constant"
	"gitlab.com/pjoc/base-service/pkg/generator"
	gc "gitlab.com/pjoc/base-service/pkg/grpc"
	"gitlab.com/pjoc/base-service/pkg/logger"
	"gitlab.com/pjoc/base-service/pkg/service"
	pb "gitlab.com/pjoc/proto/go"
	"gitlab.com/pjoc/settlement-gateway/pkg/notice"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"time"
)

const ETCD_DIR_ROOT = "/pub/pjoc/pay"

type SettlementGatewayService struct {
	*service.GatewayConfig
	OrderGenerator *generator.OrderGenerator
	*service.Service
	*gc.GrpcClientFactory
	NoticeService *notice.NoticeService
}

func (svc *SettlementGatewayService) SendNotice(ctx context.Context, notice *pb.PayNotice) (result *pb.ReturnResult, err error) {
	if notice == nil {
		err = errors.New("notice is nil")
		return
	}
	err = svc.NoticeService.Notice(notice)
	if err != nil {
		logger.Log.Errorf("Failed to send notice! error: %v", err.Error())
		return
	}

	return
}

func (svc *SettlementGatewayService) NotifyOrder(ctx context.Context, settlementPayOrder *pb.SettlementPayOrder) (response *pb.SettlementResponse, err error) {
	dbClient, err := svc.GetDatabaseClient()
	if err != nil {
		logger.Log.Errorf("Failed to get db client! error: %v", err.Error())
		return
	}

	timeoutCtx, _ := context.WithTimeout(ctx, 6*time.Second)
	order := &pb.PayOrder{}
	order.BasePayOrder = &pb.BasePayOrder{GatewayOrderId: settlementPayOrder.Order.BasePayOrder.GatewayOrderId}

	orderResponse, err := dbClient.FindPayOrder(timeoutCtx, order)
	if err != nil {
		logger.Log.Errorf("Failed to find order: %v", settlementPayOrder.Order.BasePayOrder.GatewayOrderId)
		return
	}
	if orderResponse.PayOrders == nil || len(orderResponse.PayOrders) == 0 {
		logger.Log.Errorf("Not found order! request: %v", settlementPayOrder)
		err = fmt.Errorf("failed to found order of: %v", settlementPayOrder)
		return
	}
	existsOrder := orderResponse.PayOrders[0]
	if existsOrder.OrderStatus != constant.ORDER_STATUS_SUCCESS {
		err = fmt.Errorf("order: %v is not success", existsOrder.BasePayOrder.GatewayOrderId)
		logger.Log.Error(err.Error())
		return
	}
	payNotice := svc.NoticeService.GeneratePayNotice(existsOrder)

	timeoutCtxNotice, _ := context.WithTimeout(ctx, 6*time.Second)
	result, err := dbClient.SavePayNotice(timeoutCtxNotice, payNotice)
	if err != nil {
		logger.Log.Errorf("Failed to save payNotice: %v", payNotice)
		return
	} else if result != nil && result.Code != pb.ReturnResultCode_CODE_SUCCESS {
		err = fmt.Errorf("failed to save payNotice! message: %v", result)
		logger.Log.Errorf(err.Error())
		return
	}
	returnResult, err := svc.SendNotice(ctx, payNotice)
	if err != nil {
		logger.Log.Errorf("Failed to save payNotice: %v", payNotice)
		return
	} else if returnResult != nil && returnResult.Code != pb.ReturnResultCode_CODE_SUCCESS {
		err = fmt.Errorf("failed to save payNotice! message: %v", result)
		logger.Log.Errorf(err.Error())
		return
	}
	response = &pb.SettlementResponse{}
	response.Result = returnResult
	return
}

func (svc *SettlementGatewayService) ProcessOrderSuccess(ctx context.Context, settlementPayOrder *pb.SettlementPayOrder) (response *pb.SettlementResponse, err error) {
	dbClient, err := svc.GetDatabaseClient()
	if err != nil {
		logger.Log.Errorf("Failed to get db client! error: %v", err.Error())
		return
	}
	timeoutCtx, _ := context.WithTimeout(ctx, 6*time.Second)
	orderResponse, err := dbClient.FindPayOrder(timeoutCtx, settlementPayOrder.Order)
	if err != nil {
		logger.Log.Errorf("Failed to find order: %v", settlementPayOrder.Order.BasePayOrder.GatewayOrderId)
		return
	}
	existsOrder := orderResponse.PayOrders[0]
	if existsOrder.OrderStatus == constant.ORDER_STATUS_SUCCESS {
		response = &pb.SettlementResponse{}
		response.Result = &pb.ReturnResult{Code: pb.ReturnResultCode_CODE_SUCCESS}
		return response, nil
	}

	sendStatus := constant.NOTIFY_FAILED
	// save to order ok
	timeoutSaveCtx, _ := context.WithTimeout(ctx, 6*time.Second)
	orderOk := svc.GenerateSuccessOrder(existsOrder)
	orderOk.SendNoticeStats = sendStatus

	result, e := dbClient.SavePayOrderOk(timeoutSaveCtx, orderOk);
	if e != nil {
		logger.Log.Errorf("Failed to save order ok!")
		return
	} else if result != nil && result.Code != pb.ReturnResultCode_CODE_SUCCESS {
		err = fmt.Errorf("failed to save order ok! message: %v", result)
		logger.Log.Errorf(err.Error())
		return
	}

	// notify
	timeoutNotifyCtx, _ := context.WithTimeout(ctx, 6*time.Second)
	settlementResponse, err := svc.NotifyOrder(timeoutNotifyCtx, settlementPayOrder)
	if err != nil {
		logger.Log.Errorf("Failed to notify order! error: %v order: %v", err.Error(), settlementPayOrder)
		sendStatus = constant.NOTIFY_FAILED
		return
	} else if settlementResponse.Result != nil && settlementResponse.Result.Code != pb.ReturnResultCode_CODE_SUCCESS {
		err = fmt.Errorf("failed to notify order! message: %v", settlementResponse.Result)
		logger.Log.Errorf(err.Error())
		return
	}
	sendStatus = constant.NOTIFY_SUCCESS
	orderOk.SendNoticeStats = sendStatus
	timeoutUpdateCtx, _ := context.WithTimeout(ctx, 6*time.Second)
	result, err = dbClient.UpdatePayOrderOk(timeoutUpdateCtx, orderOk)
	if err != nil {
		logger.Log.Errorf("Failed to update order ok! order: %v error: %v", orderOk, err.Error())
		return
	}

	response = &pb.SettlementResponse{}
	response.Result = &pb.ReturnResult{Code: pb.ReturnResultCode_CODE_SUCCESS}
	return response, nil
}

func (svc *SettlementGatewayService) ProcessSuccess(ctx context.Context, request *pb.SettlementRequest) (*pb.SettlementResponse, error) {
	gatewayOrderId := request.GatewayOrderId
	order := &pb.PayOrder{BasePayOrder: &pb.BasePayOrder{GatewayOrderId: gatewayOrderId}}
	settlementPayOrder := &pb.SettlementPayOrder{Order: order}
	return svc.ProcessOrderSuccess(ctx, settlementPayOrder)
}

func (svc *SettlementGatewayService) RegisterGrpc(gs *grpc.Server) {
	pb.RegisterSettlementGatewayServer(gs, svc)
}

func Init(svc *service.Service, scheduler *notice.Scheduler) *SettlementGatewayService {
	settlementGatewayService := &SettlementGatewayService{}
	settlementGatewayService.Service = svc
	flag.Parse()

	gatewayConfig := service.InitGatewayConfig(svc.EtcdPeers, ETCD_DIR_ROOT)
	settlementGatewayService.GatewayConfig = gatewayConfig

	grpcClientFactory := gc.InitGrpFactory(*svc, gatewayConfig)
	settlementGatewayService.GrpcClientFactory = grpcClientFactory

	urlGenerator := notice.NewUrlGenerator(*gatewayConfig)

	dbClient, e := grpcClientFactory.GetDatabaseClient()
	if e != nil{
		panic(e)
	}
	noticeService, err := notice.NewNoticeService(*scheduler.QueueConfig, dbClient, gatewayConfig)
	if err != nil {
		panic(err)
	}
	noticeService.UrlGenerator = urlGenerator

	settlementGatewayService.NoticeService = noticeService
	return settlementGatewayService
}

func (settlementGatewayService *SettlementGatewayService) Start() {
	settlementGatewayService.StartGrpc(settlementGatewayService.RegisterGrpc)
}
