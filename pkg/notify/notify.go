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

// Service notify service
type Service struct {
	pb.PayDatabaseServiceClient
	URLGenerator  *URLGenerator
	NotifyQueue   Queue
	configClients configclient.ConfigClients
}

// NewService create notify service
func NewService(
	config QueueConfig, dbClient pb.PayDatabaseServiceClient,
	 clients configclient.ConfigClients,
) (notifyService *Service, err error) {
	log := logger.Log()

	notifyService = &Service{}
	var queue Queue
	queue, err = InstanceQueue(config, notifyService)
	if err != nil {
		log.Errorf("Failed to init queue! error: %v", err.Error())
		return
	}
	notifyService.NotifyQueue = queue
	notifyService.PayDatabaseServiceClient = dbClient
	notifyService.configClients = clients
	notifyService.URLGenerator = NewURLGenerator(clients)
	return
}

// Notify do notify
func (svc *Service) Notify(ctx context.Context, notify *pb.PayNotify) error {
	log := logger.Log()

	err := svc.NotifyQueue.Push(ctx, *notify)
	if err != nil {
		log.Errorf("Failed to push to queue! error: %v", err)
	}
	return err
}

// GeneratePayNotify generate PayNotify
func (svc *Service) GeneratePayNotify(order *pb.PayOrder) *pb.PayNotify {
	notify := &pb.PayNotify{}

	baseOrder := order.BasePayOrder
	notify.GatewayOrderId = baseOrder.GatewayOrderId
	notify.Status = constant.OrderStatusWaiting
	notify.CreateDate = date.NowDate()
	notify.NextNotifyTime = date.NowTime()

	return notify
}

// UpdatePayNotifySuccess update notify to success
func (svc *Service) UpdatePayNotifySuccess(ctx context.Context, notify *pb.PayNotify) (err error) {
	log := logger.Log()

	notify.NotifyTime = date.NowTime()
	notify.Status = constant.OrderStatusSuccess
	notify.NotifyTime = date.NowTime()

	notifyOk := &pb.PayNotifyOk{}
	if err = copier.Copy(notifyOk, notify); err != nil {
		log.Errorf("Failed to copy instance! error: %v", err.Error())
		return
	}

	notifyOk.GatewayOrderId = notify.GatewayOrderId
	returnResult, err := svc.SavePayNotifyOk(ctx, notifyOk)
	if err != nil {
		log.Errorf(
			"Failed to update notify! orderId: %v error: %v", notify.GatewayOrderId, err.Error(),
		)
		return
	} else if returnResult.Code != pb.ReturnResultCode_CODE_SUCCESS {
		log.Errorf("Failed to update notify! orderId: %v", notify.GatewayOrderId)
	}
	return
}

// UpdatePayNotifyFail update pay notify fail
func (svc *Service) UpdatePayNotifyFail(
	ctx context.Context, notify *pb.PayNotify,
	reason error,
) (err error) {
	log := logger.Log()
	notify.Status = constant.OrderStatusFailed
	notify.ErrorMessage = reason.Error()
	notifyExpression := DefaultNotifyExpression
	if svc.GatewayConfig.NotifyConfig != nil && svc.GatewayConfig.NotifyConfig.NotifyDelaySecondExpressions != nil {
		notifyExpression = svc.GatewayConfig.NotifyConfig.NotifyDelaySecondExpressions
	}
	nextTimeStr, err := NextTimeToNotify(notify.FailTimes, notifyExpression)
	if err != nil {
		log.Errorf("Failed to build next notify time! error: %v", err.Error())
		nextTimeStr = ""
	}

	notify.FailTimes++
	notify.NextNotifyTime = nextTimeStr
	result, err := svc.UpdatePayNotify(ctx, notify)
	if err != nil {
		log.Errorf("failed to update notify: %v error: %v", err.Error())
		return
	}
	log.Infof("Update notify: %v with result: %v", notify, result)
	return
}

// SendPayNotify 发送通知
func (svc *Service) SendPayNotify(ctx context.Context, notify *pb.PayNotify) (err error) {
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

	url, form, e := svc.URLGenerator.GenerateURLByPayOrderOk(ctx, *payOrderOk)
	if e != nil {
		log.Errorf("Failed to generate url! notify: %v error: %v", notify, e.Error())
		err = e
		return
	} else if url == "" {
		errorf := fmt.Errorf("there is no notify url of order: %v ", payOrderOk)
		log.Error(errorf)
		err2 := svc.UpdatePayNotifyFail(ctx, notify, fmt.Errorf("there is no notify url of order"))
		if err2 != nil {
			log.Errorf("update db with error: %v", err2.Error())
		}
		return
	}

	resp, err := http.DefaultClient.PostForm(url, form)
	if resp != nil {
		defer func() {
			err2 := resp.Body.Close()
			if err2 != nil{
				log.Errorf("failed to close body: %v", err2.Error())
			}
		}()
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
		err = fmt.Errorf(
			"failed to notify! response is not expect value: %v factul is: %v",
			SuccessResponse, responseString,
		)
		// Save notify when fail!
		err2 := svc.UpdatePayNotifyFail(ctx, notify, err)
		if err2 != nil {
			log.Errorf("update db with error: %v", err.Error())
		}
	}
	return
}

// NextTimeToNotify next time to notify
func NextTimeToNotify(failedTimes uint32, config []int) (nextTimeStr string, err error) {
	if int(failedTimes) >= len(config) || int(failedTimes) < 0 {
		err = fmt.Errorf(
			"failed times is greater than max times! failed times: %d max failed times: %d",
			failedTimes, len(config),
		)
		return
	}
	delay := config[failedTimes]
	nextTime := time.Now().Add(time.Duration(delay) * time.Second)
	nextTimeStr = nextTime.Format(date.TimeFormat)
	return
}
