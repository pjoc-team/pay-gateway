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
func (m *MysqlQueue) Pull(ctx context.Context) (payNotifys []*pay.PayNotify, err error) {
	log := logger.ContextLog(ctx)
	payNotifyQuery := &pay.PayNotify{}
	payNotifyQuery.NextNotifyTime = date.NowTime()
	response, err := m.svc.FindPayNotifyLessThenTime(ctx, payNotifyQuery)
	if err != nil {
		log.Errorf("Failed to find notify! error: %v", err.Error())
		return
	}
	log.Infof("Found notifys: %v", response.PayNotifies)
	// 更新下一次的更新时间，防止被其他队列拉下来
	for _, notify := range response.PayNotifies {
		var nextTimeStr string
		nextTimeStr, err = NextTimeToNotify(
			notify.GetFailTimes(), m.svc.GatewayConfig.NotifyConfig.NotifyDelaySecondExpressions,
		)
		if err != nil {
			log.Errorf("Failed to get next notify time! error: %v", err.Error())
			notify.NextNotifyTime = ""
		} else {
			notify.NextNotifyTime = nextTimeStr
		}
		var result *pay.ReturnResult
		result, err = m.svc.UpdatePayNotify(ctx, notify)
		if err != nil {
			log.Errorf("Failed to update notify time! error: %v", notify)
			return
		}
		log.Infof("Update notify: %v with result: %v", notify, result)
	}
	payNotifys = response.PayNotifies
	return
}

// Push push message
func (m *MysqlQueue) Push(ctx context.Context, notify pay.PayNotify) (err error) {
	log := logger.ContextLog(ctx)

	var response *pay.PayNotifyResponse
	response, err = m.svc.FindPayNotify(ctx, &notify)
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		log.Errorf("Failed to find pay notify! error: %v", err.Error())
		return
	} else if err != nil && gorm.IsRecordNotFoundError(err) {
		var result *pay.ReturnResult
		if result, err = m.svc.SavePayNotify(ctx, &notify); err != nil {
			log.Errorf("Failed to save notify! notify: %v error: %v", notify, err.Error())
			return
		}
		log.Infof("Save notify: %v result: %v", notify, result)
		return
	}
	payNotify := response.PayNotifies[0]
	payNotify.NextNotifyTime = date.NowTime()
	var result *pay.ReturnResult
	if result, err = m.svc.UpdatePayNotify(ctx, payNotify); err != nil {
		log.Errorf("Failed to update notify! notify: %v error: %v", notify, err.Error())
		return
	}
	log.Infof("Update notify: %v result: %v", notify, result)

	return
}

// MessageSerializer serializer
func (*MysqlQueue) MessageSerializer() MessageSerializer {
	return NewJSONMessageSerializer()
}
