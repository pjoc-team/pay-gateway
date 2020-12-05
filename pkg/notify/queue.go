package notify

import (
	"context"
	"encoding/json"
	"fmt"
	pay "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
	"reflect"
)

// MessageSerializer message serializer
type MessageSerializer interface {
	// Serialize serialize message
	Serialize(pay.PayNotify) (string, error)
	// Deserialize deserialize message
	Deserialize(str string) (notify *pay.PayNotify, err error)
}

// Queue queue interface
type Queue interface {
	// Pull pull message
	Pull(ctx context.Context) ([]*pay.PayNotify, error)
	// Push push message
	Push(context.Context, pay.PayNotify) error
	// MessageSerializer serializer
	MessageSerializer() MessageSerializer
}

// QueueConfig queue config
type QueueConfig struct {
	// QueueType the type
	QueueType string `json:"queue_type"`
	// ConfigValue config value
	ConfigValue interface{} `json:"config_value"`
}

var queueConfigMap = make(map[string]interface{})
var queueTypeAndInitializeFuncMap = make(map[string]QueueInitializeFunc)
var queueTypeAndQueueMap = make(map[string]Queue)

// QueueInitializeFunc init queue func
type QueueInitializeFunc = func(QueueConfig, interface{}, *Service) (Queue, error)

// RegisterQueueType request queue type
func RegisterQueueType(
	queueType string, configValue interface{}, initializeFunc QueueInitializeFunc,
) {
	queueConfigMap[queueType] = configValue
	queueTypeAndInitializeFuncMap[queueType] = initializeFunc
}

// GetQueues all queues
func GetQueues() []Queue {
	queues := make([]Queue, 0)
	for _, queue := range queueTypeAndQueueMap {
		queues = append(queues, queue)
	}
	return queues
}

// InstanceQueue instance queue by config and svc
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
			log.Errorf(
				"Failed to unmarshal json: %v to instance type: %v, error: %v", config.ConfigValue,
				reflect.TypeOf(configInstance), err.Error(),
			)
		} else {
			log.Infof("Convert interface type to configInstance: %v", configInstance)
		}
	}
	queueFunc := queueTypeAndInitializeFuncMap[queueType]
	queue, err = queueFunc(config, configInstance, svc)
	log.Infof("init queueType: %v queue: %v error: %v", queueType, queue, err)
	return
}

// ConvertInterfaceTypeToConfigInstance convert interface to config instance
func (m *QueueConfig) ConvertInterfaceTypeToConfigInstance(configInstance interface{}) (e error) {
	bytes, e := json.Marshal(m.ConfigValue)
	if e != nil {
		return e
	}
	return json.Unmarshal(bytes, configInstance)
}
