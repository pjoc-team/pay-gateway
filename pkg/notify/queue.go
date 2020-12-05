package notify

import (
	"context"
	"encoding/json"
	"fmt"
	pay "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
	"reflect"
)

type MessageSerializer interface {
	Serialize(pay.PayNotice) (string, error)
	Deserialize(str string) (notice *pay.PayNotice, err error)
}

type Queue interface {
	Pull(ctx context.Context) ([]*pay.PayNotice, error)
	Push(context.Context, pay.PayNotice) error
	MessageSerializer() MessageSerializer
}

type QueueConfig struct {
	QueueType   string      `json:"queue_type"`
	ConfigValue interface{} `json:"config_value"`
}

var queueConfigMap = make(map[string]interface{})
var queueTypeAndInitializeFuncMap = make(map[string]QueueInitializeFunc)
var queueTypeAndQueueMap = make(map[string]Queue)

type QueueInitializeFunc = func(QueueConfig, interface{}, *Service) (Queue, error)

func RegisterQueueType(queueType string, configValue interface{}, initializeFunc QueueInitializeFunc) {
	queueConfigMap[queueType] = configValue
	queueTypeAndInitializeFuncMap[queueType] = initializeFunc
}

func GetQueues() []Queue {
	queues := make([]Queue, 0)
	for _, queue := range queueTypeAndQueueMap {
		queues = append(queues, queue)
	}
	return queues
}

func InstanceQueue(config QueueConfig, svc *Service) (queue Queue, err error) {
	log := logger.Log()
	
	queueType := config.QueueType
	configInstance, exists := queueConfigMap[queueType]
	if !exists {
		err = fmt.Errorf("could'nt found queue type: %v", queueType)
		return
	}
	if configInstance != nil && config.ConfigValue != "" {
		err := config.ConvertInterfaceTypeToConfigInstance(configInstance)
		if err != nil {
			log.Errorf("Failed to unmarshal json: %v to instance type: %v, error: %v", config.ConfigValue, reflect.TypeOf(configInstance), err.Error())
		} else {
			log.Infof("Convert interface type to configInstance: %v", configInstance)
		}
	}
	queueFunc := queueTypeAndInitializeFuncMap[queueType]
	queue, err = queueFunc(config, configInstance, svc)
	log.Infof("Init queueType: %v queue: %v error: %v", queueType, queue, err)
	return
}

func (m *QueueConfig) ConvertInterfaceTypeToConfigInstance(configInstance interface{}) (e error) {
	if bytes, e := json.Marshal(m.ConfigValue); e != nil {
		return e
	} else {
		return json.Unmarshal(bytes, configInstance)
	}
}
