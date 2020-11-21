package notify

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/pjoc-team/pay-gateway/pkg/constant"
	"github.com/pjoc-team/pay-gateway/pkg/date"
	"github.com/pjoc-team/pay-gateway/pkg/model"
	pb "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
	"io/ioutil"
	"net/http"
	"time"
)

const SUCCESS_RESPONSE = "success"

var DEFAULT_NOTICE_EXPRESSIONG = []int{30, 30, 120, 240, 480, 1200, 3600, 7200, 43200, 86400, 172800}

type NotifyService struct {
	pb.PayDatabaseServiceClient
	GatewayConfig *model.GatewayConfig
	UrlGenerator  *UrlGenerator
	NoticeQueue   Queue
}

func NewNoticeService(config QueueConfig, dbClient pb.PayDatabaseServiceClient,
	gatewayConfig *model.GatewayConfig) (noticeService *NotifyService, err error) {
	log := logger.Log()

	noticeService = &NotifyService{}
	var queue Queue
	queue, err = InstanceQueue(config, noticeService)
	if err != nil {
		log.Errorf("Failed to init queue! error: %v", err.Error())
		return
	}
	noticeService.NoticeQueue = queue
	noticeService.GatewayConfig = gatewayConfig
	noticeService.PayDatabaseServiceClient = dbClient
	noticeService.UrlGenerator = NewUrlGenerator(*gatewayConfig)
	return
}

func (svc *NotifyService) Notice(notice *pb.PayNotice) error {
	log := logger.Log()

	err := svc.NoticeQueue.Push(*notice)
	if err != nil {
		log.Errorf("Failed to push to queue! error: %v", err)
	}
	return err
}

func (svc *NotifyService) GeneratePayNotice(order *pb.PayOrder) *pb.PayNotice {
	notice := &pb.PayNotice{}

	baseOrder := order.BasePayOrder
	notice.GatewayOrderId = baseOrder.GatewayOrderId
	notice.Status = constant.ORDER_STATUS_WAITING
	notice.CreateDate = date.NowDate()
	notice.NextNotifyTime = date.NowTime()

	return notice
}

func (svc *NotifyService) UpdatePayNoticeSuccess(notice *pb.PayNotice) (err error) {
	log := logger.Log()

	notice.NoticeTime = date.NowTime()
	notice.Status = constant.ORDER_STATUS_SUCCESS
	notice.NoticeTime = date.NowTime()

	timeoutCtx, _ := context.WithTimeout(context.TODO(), 6*time.Second)

	noticeOk := &pb.PayNoticeOk{}
	if err = copier.Copy(noticeOk, notice); err != nil {
		log.Errorf("Failed to copy instance! error: %v", err.Error())
		return
	}

	noticeOk.GatewayOrderId = notice.GatewayOrderId
	returnResult, err := svc.SavePayNotifyOk(timeoutCtx, noticeOk)
	if err != nil {
		log.Errorf("Failed to update notice! orderId: %v error: %v", notice.GatewayOrderId, err.Error())
		return
	} else if returnResult.Code != pb.ReturnResultCode_CODE_SUCCESS {
		log.Errorf("Failed to update notice! orderId: %v", notice.GatewayOrderId)
	}
	return
}

func (svc *NotifyService) UpdatePayNoticeFail(notice *pb.PayNotice, reason error) (err error) {
	log := logger.Log()
	notice.Status = constant.ORDER_STATUS_FAILED
	notice.ErrorMessage = reason.Error()
	noticeExpression := DEFAULT_NOTICE_EXPRESSIONG
	if svc.GatewayConfig.NoticeConfig != nil && svc.GatewayConfig.NoticeConfig.NoticeDelaySecondExpressions != nil {
		noticeExpression = svc.GatewayConfig.NoticeConfig.NoticeDelaySecondExpressions
	}
	nextTimeStr, err := NextTimeToNotice(notice.FailTimes, noticeExpression)
	if err != nil {
		log.Errorf("Failed to build next notice time! error: %v", err.Error())
		nextTimeStr = ""
	}

	notice.FailTimes++
	notice.NextNotifyTime = nextTimeStr
	timeoutCtx, _ := context.WithTimeout(context.TODO(), 6*time.Second)
	result, err := svc.UpdatePayNotice(timeoutCtx, notice)
	if err != nil {
		return
	} else {
		log.Infof("Update notice: %v with result: %v", notice, result)
	}

	return
}

// 发送通知
func (svc *NotifyService) SendPayNotice(ctx context.Context, notice *pb.PayNotice) (err error) {
	log := logger.Log()

	payOrderOkQuery := &pb.PayOrderOk{}
	payOrderOkQuery.BasePayOrder = &pb.BasePayOrder{GatewayOrderId: notice.GatewayOrderId}
	timeoutCtx, _ := context.WithTimeout(context.TODO(), 6*time.Second)
	response, err := svc.FindPayOrderOk(timeoutCtx, payOrderOkQuery)
	if err != nil {
		log.Errorf("Failed to find order ok! error: %v", err)
		return
	}

	var payOrderOk *pb.PayOrderOk
	if len(response.PayOrderOks) > 0 {
		payOrderOk = response.PayOrderOks[0]
	} else {
		err = fmt.Errorf("not found order ok: %v", notice.GatewayOrderId)
		return
	}
	log.Infof("Found orderok: %v", payOrderOk)

	url, form, e := svc.UrlGenerator.GenerateUrlByPayOrderOk(ctx, *payOrderOk)
	if e != nil {
		log.Errorf("Failed to generate url! notice: %v error: %v", notice, e.Error())
		err = e
		return
	} else if url == "" {
		errorf := fmt.Errorf("there is no notify url of order: %v ", payOrderOk)
		log.Error(errorf)
		svc.UpdatePayNoticeFail(notice, fmt.Errorf("there is no notify url of order"))
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
	if responseString != SUCCESS_RESPONSE {
		err = fmt.Errorf("failed to notify! response is not expect value: %v factul is: %v", SUCCESS_RESPONSE, responseString)
		// Save notice when fail!
		svc.UpdatePayNoticeFail(notice, err)
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
	nextTimeStr = nextTime.Format(date.TIME_FORMAT)
	return
}
