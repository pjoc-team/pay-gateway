package notify

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/pjoc-team/pay-gateway/pkg/configclient"
	"github.com/pjoc-team/pay-gateway/pkg/constant"
	"github.com/pjoc-team/pay-gateway/pkg/date"
	"github.com/pjoc-team/pay-gateway/pkg/model"
	pb "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
	"io/ioutil"
	"net/http"
	"time"
)

// SuccessResponse return ok
const SuccessResponse = "success"

// DefaultNotifyExpression notify expression
var DefaultNotifyExpression = []int{30, 30, 120, 240, 480, 1200, 3600, 7200, 43200, 86400, 172800}

type Service struct {
	pb.PayDatabaseServiceClient
	GatewayConfig *model.GatewayConfig
	UrlGenerator  *UrlGenerator
	NotifyQueue   Queue
}

func NewService(config QueueConfig, dbClient pb.PayDatabaseServiceClient,
	gatewayConfig *model.GatewayConfig, clients configclient.ConfigClients) (noticeService *Service, err error) {
	log := logger.Log()

	noticeService = &Service{}
	var queue Queue
	queue, err = InstanceQueue(config, noticeService)
	if err != nil {
		log.Errorf("Failed to init queue! error: %v", err.Error())
		return
	}
	noticeService.NotifyQueue = queue
	noticeService.GatewayConfig = gatewayConfig
	noticeService.PayDatabaseServiceClient = dbClient
	noticeService.UrlGenerator = NewUrlGenerator(clients)
	return
}

func (svc *Service) Notify(ctx context.Context, notify *pb.PayNotice) error {
	log := logger.Log()

	err := svc.NotifyQueue.Push(ctx, *notify)
	if err != nil {
		log.Errorf("Failed to push to queue! error: %v", err)
	}
	return err
}

func (svc *Service) GeneratePayNotice(order *pb.PayOrder) *pb.PayNotice {
	notify := &pb.PayNotice{}

	baseOrder := order.BasePayOrder
	notify.GatewayOrderId = baseOrder.GatewayOrderId
	notify.Status = constant.OrderStatusWaiting
	notify.CreateDate = date.NowDate()
	notify.NextNotifyTime = date.NowTime()

	return notify
}

func (svc *Service) UpdatePayNoticeSuccess(ctx context.Context,notify *pb.PayNotice) (err error) {
	log := logger.Log()

	notify.NoticeTime = date.NowTime()
	notify.Status = constant.OrderStatusSuccess
	notify.NoticeTime = date.NowTime()

	noticeOk := &pb.PayNoticeOk{}
	if err = copier.Copy(noticeOk, notify); err != nil {
		log.Errorf("Failed to copy instance! error: %v", err.Error())
		return
	}

	noticeOk.GatewayOrderId = notify.GatewayOrderId
	returnResult, err := svc.SavePayNotifyOk(ctx, noticeOk)
	if err != nil {
		log.Errorf("Failed to update notify! orderId: %v error: %v", notify.GatewayOrderId, err.Error())
		return
	} else if returnResult.Code != pb.ReturnResultCode_CODE_SUCCESS {
		log.Errorf("Failed to update notify! orderId: %v", notify.GatewayOrderId)
	}
	return
}

func (svc *Service) UpdatePayNoticeFail(ctx context.Context,notify *pb.PayNotice,
	reason error) (err error) {
	log := logger.Log()
	notify.Status = constant.OrderStatusFailed
	notify.ErrorMessage = reason.Error()
	noticeExpression := DefaultNotifyExpression
	if svc.GatewayConfig.NoticeConfig != nil && svc.GatewayConfig.NoticeConfig.NoticeDelaySecondExpressions != nil {
		noticeExpression = svc.GatewayConfig.NoticeConfig.NoticeDelaySecondExpressions
	}
	nextTimeStr, err := NextTimeToNotice(notify.FailTimes, noticeExpression)
	if err != nil {
		log.Errorf("Failed to build next notify time! error: %v", err.Error())
		nextTimeStr = ""
	}

	notify.FailTimes++
	notify.NextNotifyTime = nextTimeStr
	result, err := svc.UpdatePayNotice(ctx, notify)
	if err != nil {
		return
	} else {
		log.Infof("Update notify: %v with result: %v", notify, result)
	}

	return
}

// 发送通知
func (svc *Service) SendPayNotice(ctx context.Context, notify *pb.PayNotice) (err error) {
	log := logger.Log()

	payOrderOkQuery := &pb.PayOrderOk{}
	payOrderOkQuery.BasePayOrder = &pb.BasePayOrder{GatewayOrderId: notify.GatewayOrderId}
	response, err := svc.FindPayOrderOk(ctx, payOrderOkQuery)
	if err != nil {
		log.Errorf("Failed to find order ok! error: %v", err)
		return
	}

	var payOrderOk *pb.PayOrderOk
	if len(response.PayOrderOks) > 0 {
		payOrderOk = response.PayOrderOks[0]
	} else {
		err = fmt.Errorf("not found order ok: %v", notify.GatewayOrderId)
		return
	}
	log.Infof("Found orderok: %v", payOrderOk)

	url, form, e := svc.UrlGenerator.GenerateUrlByPayOrderOk(ctx, *payOrderOk)
	if e != nil {
		log.Errorf("Failed to generate url! notify: %v error: %v", notify, e.Error())
		err = e
		return
	} else if url == "" {
		errorf := fmt.Errorf("there is no notify url of order: %v ", payOrderOk)
		log.Error(errorf)
		err2 := svc.UpdatePayNoticeFail(ctx, notify, fmt.Errorf("there is no notify url of order"))
		if err2 != nil{
			log.Errorf("update db with error: %v", err.Error())
		}
		return
	}

	resp, err := http.DefaultClient.PostForm(url, form)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		log.Errorf("Failed to send to url: %v and form: {%v}. Error: %v", url, form, err)
		return
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Failed to send to url: %v and form: {%v}. Error: %v", url, form, err)
		return
	}

	responseString := string(bytes)
	if responseString != SuccessResponse {
		err = fmt.Errorf("failed to notify! response is not expect value: %v factul is: %v",
			SuccessResponse, responseString)
		// Save notify when fail!
		err2 := svc.UpdatePayNoticeFail(ctx, notify, err)
		if err2 != nil{
			log.Errorf("update db with error: %v", err.Error())
		}
	}
	return
}

func NextTimeToNotice(failedTimes uint32, config []int) (nextTimeStr string, err error) {
	if int(failedTimes) >= len(config) || int(failedTimes) < 0 {
		err = fmt.Errorf("failed times is greater than max times! failed times: %d max failed times: %d", failedTimes, len(config))
		return
	}
	delay := config[failedTimes]
	nextTime := time.Now().Add(time.Duration(delay) * time.Second)
	nextTimeStr = nextTime.Format(date.TimeFormat)
	return
}
