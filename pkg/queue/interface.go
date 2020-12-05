package queue

import (
	"context"
	"fmt"
	"sync"
)

var (
	queues = make(map[string]*Info)
	lock   = sync.RWMutex{}
)

// Types 所有队列类型
func Types() (types []string) {
	for s := range queues {
		types = append(types, s)
	}
	return
}

// InstanceQueueFunc 初始化队列函数
type InstanceQueueFunc func(config interface{}) (Interface, error)

// Info 队列信息
type Info struct {
	InstanceQueueFunc
	ConfigDemo interface{}
}

// Interface 实际队列的实现
type Interface interface {
	// Consume consume messages.
	//
	// messages: is the channel of message need to write.
	ConsumeTopics(ctx context.Context, topics string, messages chan<- *Message) (err error)

	// Producer
	//
	// messages: is the channel of message need to produce.
	ProduceTopic(ctx context.Context, topic string, messages <-chan *Message) (err error)

	// Push push message top topic.
	Push(ctx context.Context, topic string, message *Message) (err error)

	// Stop
	Stop(context.Context) error
}

// PartitionConsumer 分区消费者
type PartitionConsumer func(ctx context.Context, message <-chan *Message, ack chan<- *Message) error

// Queue 队列
type Queue interface {

	// PartitionConsume 分区消费
	PartitionConsume(ctx context.Context, topics string, consumers []PartitionConsumer) error

	// Consume consume messages.
	//
	// messages: is the channel of message need to write.
	//
	// Ack: processed message channel of this handler.
	ConsumeTopics(ctx context.Context, topics string, messages chan<- *Message, ack <-chan *Message) (err error)

	// Producer
	//
	// messages: is the channel of message need to produce.
	ProduceTopic(ctx context.Context, topic string, messages <-chan *Message) (err error)

	// Stop
	Stop(context.Context) error
}

// Register 注册队列
func Register(queueType string, queue InstanceQueueFunc, configDemo interface{}) error {
	lock.Lock()
	defer lock.Unlock()
	exists, ok := queues[queueType]
	if ok {
		err := fmt.Errorf("type: %v already exists: %v", queueType, exists)
		return err
	}
	info := &Info{
		InstanceQueueFunc: queue,
		ConfigDemo:        configDemo,
	}
	queues[queueType] = info
	return nil
}

// GetQueue 获取对应类型的信息
func GetQueue(queueType string) (*Info, error) {
	lock.RLock()
	defer lock.RUnlock()
	registerInfo := queues[queueType]
	if registerInfo == nil {
		err := fmt.Errorf("not found type: %v", queueType)
		return nil, err
	}
	return registerInfo, nil
}
