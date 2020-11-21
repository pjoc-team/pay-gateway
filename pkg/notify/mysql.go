package notify

import (
	"context"
	"github.com/jinzhu/gorm"
	"github.com/pjoc-team/pay-gateway/pkg/date"
	pay "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
	"sync"
	"time"
)

const (
	QUEUE_TYPE_MYSQL = "mysql"
)

func init() {
	RegisterQueueType(QUEUE_TYPE_MYSQL, &MySQLConfig{}, instanceFunc)
}

func instanceFunc(queueConfig QueueConfig, config interface{}, svc *NotifyService) (Queue, error) {
	log := logger.Log()
	queue := &MysqlQueue{}
	queue.svc = svc
	log.Infof("svc: %v", svc)
	queue.config = config.(*MySQLConfig)
	queue.queueConfig = queueConfig
	return queue, nil
}

type MySQLConfig struct {
	//MySQLConnectionUrl string `json:"mysql_connection_url"`
	//DatabaseName       string `json:"database_name"`
	//TableName          string `json:"table_name"`
	//TimeColumnName     string `json:"time_column_name"`
}

type MysqlQueue struct {
	config      *MySQLConfig
	svc         *NotifyService
	queueConfig QueueConfig
	sync.Mutex
}

func (m *MysqlQueue) Pull() (payNotices []*pay.PayNotice, err error) {
	log := logger.Log()
	timeoutCtx, _ := context.WithTimeout(context.TODO(), 6*time.Second)
	payNoticeQuery := &pay.PayNotice{}
	payNoticeQuery.NextNotifyTime = date.NowTime()
	response, err := m.svc.FindPayNoticeLessThenTime(timeoutCtx, payNoticeQuery)
	if err != nil {
		log.Errorf("Failed to find notice! error: %v", err.Error())
		return
	} else {
		log.Infof("Found notices: %v", response.PayNotices)
	}
	// 更新下一次的更新时间，防止被其他队列拉下来
	for _, notice := range response.PayNotices {
		var nextTimeStr string
		nextTimeStr, err = NextTimeToNotice(notice.GetFailTimes(), m.svc.GatewayConfig.NoticeConfig.NoticeDelaySecondExpressions)
		if err != nil {
			log.Errorf("Failed to get next notice time! error: %v", err.Error())
			notice.NextNotifyTime = ""
		} else {
			notice.NextNotifyTime = nextTimeStr
		}
		timeout, _ := context.WithTimeout(context.TODO(), 6*time.Second)
		var result *pay.ReturnResult
		result, err = m.svc.UpdatePayNotice(timeout, notice)
		if err != nil {
			log.Errorf("Failed to update notice time! error: %v", notice)
			return
		} else {
			log.Infof("Update notice: %v with result: %v", notice, result)
		}
	}
	payNotices = response.PayNotices
	return
}

func (m *MysqlQueue) Push(notice pay.PayNotice) (err error) {
	log := logger.Log()

	timeoutCtx, _ := context.WithTimeout(context.TODO(), 6*time.Second)

	var response *pay.PayNoticeResponse
	response, err = m.svc.FindPayNotice(timeoutCtx, &notice)
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		log.Errorf("Failed to find pay notice! error: %v", err.Error())
		return
	} else if err != nil && gorm.IsRecordNotFoundError(err) {
		timeoutCtx2, _ := context.WithTimeout(context.TODO(), 6*time.Second)
		var result *pay.ReturnResult
		if result, err = m.svc.SavePayNotice(timeoutCtx2, &notice); err != nil {
			log.Errorf("Failed to save notice! notice: %v error: %v", notice, err.Error())
			return
		} else {
			log.Infof("Save notice: %v result: %v", notice, result)
			return
		}
	}
	payNotice := response.PayNotices[0]
	payNotice.NextNotifyTime = date.NowTime()
	timeoutCtx2, _ := context.WithTimeout(context.TODO(), 6*time.Second)
	var result *pay.ReturnResult
	if result, err = m.svc.UpdatePayNotice(timeoutCtx2, payNotice); err != nil {
		log.Errorf("Failed to update notice! notice: %v error: %v", notice, err.Error())
		return
	} else {
		log.Infof("Update notice: %v result: %v", notice, result)
	}

	return
}

func (*MysqlQueue) MessageSerializer() MessageSerializer {
	return NewJsonMessageSerializer()
}
