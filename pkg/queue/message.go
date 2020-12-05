package queue

import (
	"context"
	"fmt"
)

// Message 队列消息
type Message struct {
	ID       string
	Row      []byte
	Ack      func(ctx context.Context) `json:"-"` // the func to mark message as processed
	Rollback func(ctx context.Context) `json:"-"` // Rollback to the input. for example, kafka is produce the message back to kafka
	Context  context.Context           `json:"-"` // Context of the mq lifetime
}

// NewMessage 创建消息
func NewMessage(ctx context.Context, id string, row []byte, ack func(ctx context.Context), rollback func(ctx context.Context)) *Message {
	m := &Message{
		ID:       id,
		Row:      row,
		Ack:      ack,
		Rollback: rollback,
		Context:  ctx,
	}
	return m
}

func (m *Message) String() string {
	return fmt.Sprintf("id: %v row: %v", m.ID, m.Row)
}
