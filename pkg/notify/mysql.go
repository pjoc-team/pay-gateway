package notify

import (
	"context"
	"github.com/jinzhu/gorm"
	"github.com/pjoc-team/pay-gateway/pkg/date"
	pay "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
	"sync"
)

const (
	// QueueTypeMysql mysql queue
	QueueTypeMysql = "mysql"
)

func init() {
	RegisterQueueType(QueueTypeMysql, &MySQLConfig{}, instanceFunc)
}

func instanceFunc(queueConfig QueueConfig, config interface{}, svc *Service) (Queue, error) {
	log := logger.Log()
	queue := &MysqlQueue{}
	queue.svc = svc
	log.Infof("svc: %v", svc)
	queue.config = config.(*MySQLConfig)
	queue.queueConfig = queueConfig
	return queue, nil
}

// MySQLConfig mysql queue config
type MySQLConfig struct {
	// MySQLConnectionUrl string `json:"mysql_connection_url"`
	// DatabaseName       string `json:"database_name"`
	// TableName          string `json:"table_name"`
	// TimeColumnName     string `json:"time_column_name"`
}

// MysqlQueue queue type: mysql
type MysqlQueue struct {
	config      *MySQLConfig
	svc         *Service
	queueConfig QueueConfig
	sync.Mutex
}

// Pull pull message
func (m *MysqlQueue) Pull(ctx context.Context) (payNotices []*pay.PayNotice, err error) {
	log := logger.ContextLog(ctx)
	payNoticeQuery := &pay.PayNotice{}
	payNoticeQuery.NextNotifyTime = date.NowTime()
	response, err := m.svc.FindPayNoticeLessThenTime(ctx, payNoticeQuery)
	if err != nil {
		log.Errorf("Failed to find notify! error: %v", err.Error())
		return
	}
	log.Infof("Found notices: %v", response.PayNotices)
	// 更新下一次的更新时间，防止被其他队列拉下来
	for _, notice := range response.PayNotices {
		var nextTimeStr string
		nextTimeStr, err = NextTimeToNotice(
			notice.GetFailTimes(), m.svc.GatewayConfig.NoticeConfig.NoticeDelaySecondExpressions,
		)
		if err != nil {
			log.Errorf("Failed to get next notify time! error: %v", err.Error())
			notice.NextNotifyTime = ""
		} else {
			notice.NextNotifyTime = nextTimeStr
		}
		var result *pay.ReturnResult
		result, err = m.svc.UpdatePayNotice(ctx, notice)
		if err != nil {
			log.Errorf("Failed to update notify time! error: %v", notice)
			return
		}
		log.Infof("Update notify: %v with result: %v", notice, result)
	}
	payNotices = response.PayNotices
	return
}

// Push push message
func (m *MysqlQueue) Push(ctx context.Context, notice pay.PayNotice) (err error) {
	log := logger.ContextLog(ctx)

	var response *pay.PayNoticeResponse
	response, err = m.svc.FindPayNotice(ctx, &notice)
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		log.Errorf("Failed to find pay notify! error: %v", err.Error())
		return
	} else if err != nil && gorm.IsRecordNotFoundError(err) {
		var result *pay.ReturnResult
		if result, err = m.svc.SavePayNotice(ctx, &notice); err != nil {
			log.Errorf("Failed to save notify! notify: %v error: %v", notice, err.Error())
			return
		}
		log.Infof("Save notify: %v result: %v", notice, result)
		return
	}
	payNotice := response.PayNotices[0]
	payNotice.NextNotifyTime = date.NowTime()
	var result *pay.ReturnResult
	if result, err = m.svc.UpdatePayNotice(ctx, payNotice); err != nil {
		log.Errorf("Failed to update notify! notify: %v error: %v", notice, err.Error())
		return
	}
	log.Infof("Update notify: %v result: %v", notice, result)

	return
}

// MessageSerializer serializer
func (*MysqlQueue) MessageSerializer() MessageSerializer {
	return NewJSONMessageSerializer()
}
