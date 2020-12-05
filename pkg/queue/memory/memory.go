package memory

import (
	"context"
	"github.com/pjoc-team/pay-gateway/pkg/queue"
	"github.com/pjoc-team/tracing/logger"
	"sync"

)

// Type 内存型MQ
const Type = "memory"

func init() {
	log := logger.Log()

	err := queue.Register(Type, func(config interface{}) (queue.Interface, error) {
		return &q{
			topicAndMessage: make(map[string]chan *queue.Message),
		}, nil
	}, nil)
	if err != nil {
		log.Fatalf("failed to register memory queue", err.Error())
	}
}

type q struct {
	topicAndMessage map[string]chan *queue.Message
	locker sync.Mutex
}

func (q *q) ConsumeTopics(ctx context.Context, topics string,
	messages chan<- *queue.Message) (err error) {
	q.locker.Lock()
	defer q.locker.Unlock()
	log := logger.ContextLog(ctx)

	topicArray := queue.ParseBrokers(topics)
	for _, s := range topicArray {
		message, ok := q.topicAndMessage[s]
		if !ok {
			message = make(chan *queue.Message, 1024)
			q.topicAndMessage[s] = message
		}
		go func() {
			for {
				select {
				case <-ctx.Done():
					log.Warnf("stopped ack")
					return
				case m := <-message:
					messages <- m
				}
			}
		}()
	}
	return nil
}

func (q *q) Push(ctx context.Context, topic string, message *queue.Message) (err error) {
	mc, ok := q.topicAndMessage[topic]
	if !ok {
		mc = make(chan *queue.Message, 1024)
		q.topicAndMessage[topic] = mc
	}
	select {
	case mc <- message:
	case <-ctx.Done():
	}
	return
}

// Ack ack message
func (q *q) Ack(ctx context.Context, ack <-chan *queue.Message) {
	log := logger.ContextLog(ctx)

	select {
	case <-ctx.Done():
		log.Warnf("stopped ack")
		return
	case a := <-ack:
		if a.Ack != nil {
			a.Ack(ctx)
		}
	}
}

func (q *q) ProduceTopic(ctx context.Context, topic string, messages <-chan *queue.Message) (err error) {
	q.locker.Lock()
	defer q.locker.Unlock()
	log := logger.ContextLog(ctx)

	message, ok := q.topicAndMessage[topic]
	if !ok {
		message = make(chan *queue.Message, 1024)
		q.topicAndMessage[topic] = message
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Warnf("stopped ack")
				return
			case m := <-messages:
				message <- m
			}
		}
	}()
	return
}

func (q *q) Stop(ctx context.Context) error {
	return nil
}
