package queue

import (
	"encoding/json"
	"fmt"
	reflect2 "github.com/pjoc-team/pay-gateway/pkg/util/reflect"
	"github.com/pjoc-team/tracing/logger"
	"reflect"
)

// Config 队列配置
type Config struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

// // Init 从队列配置初始化
// func (config *Config) Init() (Interface, error) {
//	itf, err := config.GetQueue()
//	if err != nil{
//		return nil, err
//	}
//	return itf, nil
// }

// GetQueue 从队列配置初始化queue
func (c *Config) GetQueue() (Interface, error) {
	log := logger.Log()

	info, err := GetQueue(c.Type)
	if err != nil {
		return nil, err
	}

	var config interface{}
	if info.ConfigDemo != nil {
		t := reflect.TypeOf(info.ConfigDemo)
		config = reflect2.CloneNew(info.ConfigDemo)

		marshal, err := json.Marshal(c.Config)
		if err != nil {
			err := fmt.Errorf("failed to marshal config: %v to json: %v", c, err.Error())
			return nil, err
		}
		err = json.Unmarshal(marshal, config)
		if err != nil {
			err := fmt.Errorf("failed to unmarshal config: %v to json: %v", c, err.Error())
			return nil, err
		}

		switch t.Kind() {
		case reflect.Ptr:
		default:
			config = reflect.ValueOf(config).Elem().Interface()
		}
	}

	q, err := info.InstanceQueueFunc(config)
	if err != nil {
		demoJSON := ""
		cc, ej := json.Marshal(info.ConfigDemo)
		if ej == nil {
			demoJSON = string(cc)
		}
		log.Errorf("not init queue type: %v, error: %v, a demo config is: %v", c.Type, err.Error(), demoJSON)
		err = fmt.Errorf("failed to init queue of type: %v error: %v", c.Type, err.Error())
		return nil, err
	}

	log.Infof("succeed init queue: %v by config: %v", c.Type, config)

	return q, nil
}
