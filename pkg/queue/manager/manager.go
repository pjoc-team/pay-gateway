package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/blademainer/commons/pkg/recoverable"
	"github.com/blademainer/commons/pkg/retryer"
	"github.com/pjoc-team/pay-gateway/pkg/queue"
	"github.com/pjoc-team/tracing/logger"
	"github.com/pjoc-team/tracing/tracing"
	"github.com/pjoc-team/tracing/tracingmq"
	"hash/fnv"
	"sync"
	"time"
)

type topicMessage struct {
	topic string
	mc    <-chan *queue.Message
}

type queueManager struct {
	locker      sync.RWMutex
	initialized bool // 标记队列是否已经初始化
	config      *queue.Config
	queue       queue.Interface
	producers   []*topicMessage
	retryer     retryer.Retryer
}

func partition(id string, num int) int {
	h := fnv.New32a()
	h.Write([]byte(id))
	sum32 := int(h.Sum32())
	if sum32 < 0 {
		sum32 = -sum32
	}
	return sum32 % num
}

// PartitionConsume 分区消费
func (q *queueManager) PartitionConsume(ctx context.Context, topics string, consumers []queue.PartitionConsumer) error {
	l := len(consumers)
	if l <= 0 {
		return fmt.Errorf("consumer must greater than 0")
	}
	log := logger.ContextLog(ctx)
	// 对应mq的队列≤
	messages := make(chan *queue.Message, l)
	// 分区消费者对应的channel，每个consumer都有一个channel保证顺序消费
	consumerMessages := make([]chan *queue.Message, l)
	ack := make(chan *queue.Message, l)
	err := q.ConsumeTopics(ctx, topics, messages, ack)
	if err != nil {
		return err
	}
	// 分别启动n个consumer
	for i, consumer := range consumers {
		mc := make(chan *queue.Message, l)
		consumerMessages[i] = mc
		go func(c queue.PartitionConsumer, messageChan chan *queue.Message) {
			err := c(ctx, messageChan, ack)
			if err != nil {
				log.Errorf("failed to consumer message: %v", err.Error())
			}
		}(consumer, mc)
	}
	// 消费数据并根据ID取模分配个对应的consumer
	go func() {
		log.Infof("starting handle partition consume of topic: %v consumer size: %v", topics, len(consumers))
		for {
			log.Infof("messages: %v", messages)
			select {
			case <-ctx.Done():
				log.Fatal("PartitionConsumer closed...")
				return
			case m, ok := <-messages:
				if !ok {
					log.Errorf("consumer's channel is closed!")
					return
				}
				index := partition(m.ID, l)
				f := consumerMessages[index]
				if log.IsDebugEnabled() {
					log.Debugf("push message: %v to consumer index: %v", m.ID, index)
				}
				f <- m
			}
		}
	}()
	return nil
}

func (q *queueManager) ConsumeTopics(ctx context.Context, topics string, messages chan<- *queue.Message, ack <-chan *queue.Message) (err error) {
	if topics == "" || messages == nil || ack == nil {
		err = fmt.Errorf("topics,messages and ack can't be null")
		return
	}
	// retry func
	pf := func(context.Context) error {
		q.locker.Lock()
		defer q.locker.Unlock()
		if !q.initialized {
			return &retryer.RetryError{InnerError: &NotInitializedError{Message: "not initialized"}}
		}
		go q.ack(ctx, ack)
		tracingMessages := make(chan *queue.Message)
		// 解析信息，生成context
		go q.deliverConsumerTracingMessages(ctx, topics, tracingMessages, messages)
		err = q.queue.ConsumeTopics(ctx, topics, tracingMessages)
		return err
	}

	if !q.initialized {
		err := q.retryer.Invoke(pf)
		if !retryer.IsRetryError(err) {
			return err
		}
		return nil
	}
	return pf(ctx)

}

func marshalTracingMessage(m *queue.Message, topics string) (*queue.Message, error) {
	if m == nil {
		return nil, nil
	}
	if m.Context == nil {
		m.Context = context.Background()
	}
	traceMqData := tracingmq.TracingMqProducer(m.Context, topics, m.Row)
	data, err := json.Marshal(traceMqData)
	if err != nil {
		return nil, err
	}
	log := logger.ContextLog(m.Context)
	if log.IsDebugEnabled() {
		log.Debugf("marshal data: %v to tracing json: %v", string(m.Row), string(data))
	}
	result := &queue.Message{}
	result.Row = data
	result.Rollback = m.Rollback
	result.Ack = m.Ack
	result.ID = m.ID
	return result, nil
}

func unmarshalTracingMessage(m *queue.Message, topics string) (*queue.Message, error) {
	data := &tracingmq.TraceMqData{}
	err := json.Unmarshal(m.Row, data)
	if err != nil {
		return nil, err
	}
	ctx := tracing.BuildContextByCarrier(data.Carriers, "queue_consumer", topics)
	result := &queue.Message{}
	result.Row = data.Data
	result.ID = m.ID
	result.Ack = m.Ack
	result.Rollback = m.Rollback
	result.Context = ctx
	log := logger.ContextLog(ctx)
	if log.IsDebugEnabled() {
		log.Debugf("unmarshal tracing json: %v to data: %v", string(data.Data), string(m.Row))
	}

	return result, nil
}

func (q *queueManager) deliverConsumerTracingMessages(ctx context.Context, topics string, in <-chan *queue.Message, out chan<- *queue.Message) {
	log := logger.ContextLog(ctx)
	for {
		select {
		case m := <-in:
			message, err2 := unmarshalTracingMessage(m, topics)
			if err2 != nil {
				log.Errorf("failed to unmarshal message: %v error: %v", string(m.Row))
				continue
			}
			if log.IsDebugEnabled() {
				log := logger.ContextLog(message.Context)
				log.Debugf("consume message: %v", string(message.Row))
			}
			out <- message
		case <-ctx.Done():
			log.Warnf("deliverConsumerTracingMessages done.")
			return
		}
	}
}

func (q *queueManager) deliverProducerTracingMessages(ctx context.Context, topics string, in <-chan *queue.Message, out chan<- *queue.Message) {
	log := logger.Log()
	for {
		select {
		case <-ctx.Done():
			log.Warnf("context done.")
			return
		case m := <-in:
			message, err2 := marshalTracingMessage(m, topics)
			if err2 != nil {
				log.Errorf("failed to marshal message: %v error: %v", string(m.Row))
				continue
			}
			if log.IsDebugEnabled() {
				log := logger.ContextLog(m.Context)
				log.Debugf("produce message: %v", string(message.Row))
			}
			out <- message
		}
	}
}

func (q *queueManager) ack(ctx context.Context, ack <-chan *queue.Message) {
	log := logger.ContextLog(ctx)
	defer func() {
		log.Warnf("ack is stopped")
	}()
	for {
		select {
		case <-ctx.Done():
			log.Warnf("stopped ack")
			return
		case a, ok := <-ack:
			if !ok {
				log.Warnf("ack channel is closed")
				return
			}
			log.Debugf("ack: %v", string(a.Row))
			if a.Ack != nil {
				recoverable.WithRecover(func() {
					a.Ack(ctx)
				})
			}
		}
	}
}

func (q *queueManager) ProduceTopic(ctx context.Context, topic string, messages <-chan *queue.Message) (err error) {
	if messages == nil || topic == "" {
		err = fmt.Errorf("neither messages nor topic can be null")
		return
	}
	log := logger.Log()

	// retry func
	pf := func(context.Context) error {
		q.locker.Lock()
		defer q.locker.Unlock()
		if !q.initialized {
			return &retryer.RetryError{InnerError: &NotInitializedError{Message: "not initialized"}}
		}
		log.Warnf("retry produce topic: %#v", topic)

		tracingMessages := make(chan *queue.Message)
		// 编码消息，将tracing信息注入进去
		go q.deliverProducerTracingMessages(ctx, topic, messages, tracingMessages)

		q.producers = append(q.producers, &topicMessage{topic: topic, mc: messages}, &topicMessage{topic: topic, mc: tracingMessages})

		return q.queue.ProduceTopic(ctx, topic, tracingMessages)
	}

	if !q.initialized {
		err := q.retryer.Invoke(pf)
		if !retryer.IsRetryError(err) {
			return err
		}
		return nil
	}
	return pf(ctx)
}

func (q *queueManager) Stop(ctx context.Context) error {
	q.locker.Lock()
	defer q.locker.Unlock()
	if !q.initialized {
		return nil
	}
	log := logger.ContextLog(ctx)
	log.Warnf("begin push least messages")
	for i := 0; i < len(q.producers); i++ {
		p := q.producers[i]
		log.Warnf("begin push least message topic: %v messages: %v", p.topic, len(p.mc))
		q.pushLeastMessages(ctx, p)
	}
	q.queue.Stop(ctx)
	return nil
}

// pushLeastMessages 推送线程内剩余的消息
func (q *queueManager) pushLeastMessages(ctx context.Context, p *topicMessage) {
	log := logger.ContextLog(ctx)

	for i := 0; i < len(p.mc); i++ {
		m, ok := <-p.mc
		if !ok {
			break
		}
		err := q.queue.Push(ctx, p.topic, m)
		if err != nil {
			log.Errorf("failed to handle least message: %v error: %v", string(m.Row), err.Error())
		} else {
			log.Infof("succeed push message: %v", string(m.Row))
		}
	}
}

// NewQueue 从队列配置初始化
func NewQueue(config *queue.Config) (queue.Queue, error) {
	log := logger.Log()

	q := &queueManager{}
	q.config = config
	itf, err := config.GetQueue()
	if err != nil {
		// 初始化重试器
		growthRetryer, err2 := retryer.NewRetryer(retryer.NewDefaultDoubleGrowthRateRetryStrategy(), 1000, 65535, 10*time.Second, 10*time.Second, retryer.DiscardStrategyEarliest)

		q.initialized = false
		var f = func(ctx context.Context) error {
			log.Warnf("retry init queue")
			q.locker.Lock()
			defer q.locker.Unlock()
			itf, err := config.GetQueue()
			if err != nil {
				return &retryer.RetryError{InnerError: err}
			}
			q.queue = itf
			q.initialized = true
			return nil
		}

		if err2 != nil {
			log.Errorf("failed to init retryer with error: %v", err2.Error())
			return nil, err
		}
		q.retryer = growthRetryer
		q.initialized = false
		err2 = growthRetryer.Invoke(f)
		if !retryer.IsRetryError(err2) {
			log.Errorf("failed to init retryer with error: %v", err2.Error())
			return nil, err
		}
	} else {
		q.queue = itf
		q.initialized = true
	}

	return q, nil
}
