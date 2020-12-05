package kafka

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/pjoc-team/pay-gateway/pkg/queue"
	"github.com/pjoc-team/tracing/logger"
	"gopkg.in/go-playground/validator.v9"
	"reflect"
)

const (
	// Type kafka类型
	Type = "kafka"
	// OffsetNewest stands for the log head offset, i.e. the offset that will be
	// assigned to the next message that will be produced to the partition. You
	// can send this to a client's GetOffset method to get this offset, or when
	// calling ConsumePartition to start consuming new messages.
	OffsetNewest int64 = -1
	// OffsetOldest stands for the oldest offset available on the broker for a
	// partition. You can send this to a client's GetOffset method to get this
	// offset, or when calling ConsumePartition to start consuming from the
	// oldest offset that is still available on the broker.
	OffsetOldest int64 = -2
)

var offsetConfigMap = make(map[string]int64)

func init() {
	log := logger.Log()

	err := queue.Register(
		Type,
		instanceFunc,
		&Config{
			Brokers:       "127.0.0.1:9092",
			Group:         "test",
			Version:       "1.1.1",
			OffsetInitial: "OffsetOldest",
		},
	)

	if err != nil {
		log.Fatalf("failed to register kafka input, error: %v", err.Error())
	}

	offsetConfigMap["OffsetNewest"] = OffsetNewest
	offsetConfigMap["OffsetOldest"] = OffsetOldest
}

func instanceFunc(ic interface{}) (i queue.Interface, err error) {
	log := logger.Log()

	c, ok := ic.(*Config)
	if !ok {
		err = fmt.Errorf("could'nt convert type: %v to kafka config", reflect.TypeOf(ic))
		log.Errorf(err.Error())
		return
	}
	var validate = validator.New()
	err = validate.Struct(c)
	if err != nil {
		log.Errorf("failed to validate config: %v, error: %v", ic, err.Error())
		return
	}

	k := &kafka{config: c}

	// init consumer and Producer
	config := sarama.NewConfig()

	// settings logger
	sarama.Logger = log

	// config.consumer.Return.Errors = true
	if c.OffsetInitial != "" {
		offsetInitial, ok := offsetConfigMap[c.OffsetInitial]
		if ok {
			config.Consumer.Offsets.Initial = offsetInitial
		} else {
			config.Consumer.Offsets.Initial = sarama.OffsetNewest
		}
	} else {
		config.Consumer.Offsets.Initial = sarama.OffsetNewest
	}
	brokers := queue.ParseBrokers(c.Brokers)
	if c.Version != "" {
		config.Version, err = sarama.ParseKafkaVersion(c.Version)
		if err != nil {
			err = fmt.Errorf("could'nt parse KafkaVersion: %v to kafka config, error: %v", c.Version, err.Error())
			return nil, err
		}
	} else {
		config.Version = sarama.V0_10_2_0
	}
	config.Producer.Return.Successes = true
	config.Consumer.Return.Errors = true

	k.consumerGroupFunc = func() (sarama.ConsumerGroup, error) {
		consumer, err := sarama.NewConsumerGroup(brokers, c.Group, config)
		return consumer, err
	}

	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, config)

	if err != nil {
		return nil, err
	}
	k.producer = producer

	return k, nil
}

// Config kafka配置
type Config struct {
	Brokers string `json:"brokers" validate:"required"`
	Group         string `json:"group" validate:"required"`
	Version       string `json:"version"`
	OffsetInitial string `json:"offset_initial"`
}

type kafka struct {
	config            *Config
	kafkaConfig       sarama.Config
	consumerGroupFunc func() (sarama.ConsumerGroup, error)
	producer          sarama.SyncProducer
}

func (k *kafka) Push(ctx context.Context, topic string, message *queue.Message) (err error) {
	log := logger.ContextLog(ctx)

	row := message.Row
	producerMessage := &sarama.ProducerMessage{}
	producerMessage.Key = sarama.StringEncoder(message.ID)
	producerMessage.Value = sarama.ByteEncoder(row)
	producerMessage.Topic = topic
	partition, offset, err := k.producer.SendMessage(producerMessage)
	if err != nil {
		log.Errorf("failed to send message to kafka: %s, error: %v", message, err)
	} else if log.IsDebugEnabled() {
		log.Debugf("sent message: %v, partition: %v offset: %v", string(row), partition, offset)
	}
	return err
}

type kafkaConsumer struct {
	mc          chan<- *queue.Message
	ctx         context.Context
	kafkaConfig *Config
}

func (k *kafkaConsumer) Setup(session sarama.ConsumerGroupSession) error {
	log := logger.ContextLog(k.ctx)
	log.Infof("setup... %#v", session.Claims())
	return nil
}

func (k *kafkaConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (k *kafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	log := logger.ContextLog(k.ctx)
	log.Infof("begin consume topic: %v", claim.Topic())

	for {
		select {
		case m, ok := <-claim.Messages():
			if !ok {
				log.Infof("Consumer closed... stream: %v", k)
				return nil
			}
			value := m.Value
			if log.IsDebugEnabled() {
				log.Debugf("receive value: %v", string(value))
			}
			msg := &queue.Message{
				ID:  string(m.Key),
				Row: value,
			}
			msg.Ack = func(ctx context.Context) {
				session.MarkMessage(m, "")
			}
			select {
			case k.mc <- msg:
				log.Debugf("succeed to push message: %v", string(msg.Row))
				//default:
				//log.Errorf("msg chan is full!")
			}

		case <-k.ctx.Done():
			log.Warnf("stream: %v closed", k)
			return nil
		}
	}
}

// NewConsumer create consumer
func (k *kafka) NewConsumer(ctx context.Context, kafkaConfig *Config, mc chan<- *queue.Message) *kafkaConsumer {
	consumer := &kafkaConsumer{
		mc:          mc,
		ctx:         ctx,
		kafkaConfig: kafkaConfig,
	}
	return consumer
}

// Consume consume messages.
//
// messages: is the channel of message need to write.
//
// ack: processed message channel of this handler.
func (k *kafka) ConsumeTopics(ctx context.Context, topics string, messages chan<- *queue.Message) (err error) {
	topicArray := queue.ParseBrokers(topics)
	log := logger.ContextLog(ctx)
	log.Infof("begin consume topic: %v", topics)
	go k.consume(ctx, topicArray, messages)
	return nil
}

func (k *kafka) consume(ctx context.Context, topicArray []string, messages chan<- *queue.Message) (err error) {
	log := logger.ContextLog(ctx)
	for {
		log.Infof("begin consume topic: %#v", topicArray)
		select {
		case <-ctx.Done():
			log.Warnf("consume done.")
			return
		default:
			consumer := k.NewConsumer(ctx, k.config, messages)
			consumerGroup, err := k.consumerGroupFunc()
			if err != nil {
				log.Errorf("failed to init consumerGroup with error: %v", err.Error())
				continue
			}
			err = consumerGroup.Consume(ctx, topicArray, consumer)
			if err != nil {
				log.Errorf("failed to consume, config: %v error: %v", k.config, err.Error())
			}
		}

	}
}

func (k *kafka) ack(ctx context.Context, ack <-chan *queue.Message) {
	log := logger.ContextLog(ctx)
	for {
		select {
		case <-ctx.Done():
			log.Warnf("stopped ack")
			return
		case a := <-ack:
			log.Debugf("ack: %v", a)
			a.Ack(ctx)
		}
	}
}

// Producer
//
// messages: is the channel of message need to produce.
func (k *kafka) ProduceTopic(ctx context.Context, topic string, messages <-chan *queue.Message) error {
	log := logger.ContextLog(ctx)
	go func() {
		for {
			select {
			case m, ok := <-messages:
				if !ok {
					return
				}
				err := k.Push(ctx, topic, m)
				if err != nil {
					log.Errorf("failed to push message: %v error: %v", m, err)
				}
			case <-ctx.Done():
				log.Warnf("stopped producer")
				return
			}
		}
	}()
	return nil
}

func (k *kafka) Stop(context.Context) error {
	return nil
}
