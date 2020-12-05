package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pjoc-team/pay-gateway/pkg/queue"
	_ "github.com/pjoc-team/pay-gateway/pkg/queue/kafka"
	_ "github.com/pjoc-team/pay-gateway/pkg/queue/memory"
	"github.com/pjoc-team/tracing/logger"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
)

func ExampleNewQueue() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cnf := `{"type":"kafka","config":{"brokers":"127.0.0.1:9092","group":"pushLeastMessages","version":"1.1.1","offset_initial":"OffsetOldest"}}`
	c := &queue.Config{}
	err := json.Unmarshal([]byte(cnf), c)
	if err != nil {
		panic(err)
	}

	q, err := NewQueue(c)
	if err != nil {
		panic(err)
	}

	p := make(chan *queue.Message, 1024)

	err = q.ProduceTopic(ctx, "pushLeastMessages", p)
	if err != nil {
		panic(err)
	}

	cc := make(chan *queue.Message, 1024)
	ack := make(chan *queue.Message, 1024)

	for i := 0; i < 2; i++ {
		err = q.ConsumeTopics(ctx, "pushLeastMessages", cc, ack)
		if err != nil {
			panic(err)
		}

	}

	size := 100
	wg := sync.WaitGroup{}
	wg.Add(size)
	go func() {
		for {
			select {
			case m := <-cc:
				fmt.Println(string(m.Row))
				ack <- m
				wg.Done()
			case <-ctx.Done():
				fmt.Println("done...")
				return
			}
		}
	}()

	for i := 0; i < size; i++ {
		m := &queue.Message{
			ID:  fmt.Sprintf("%d", i),
			Row: []byte(fmt.Sprintf("hello%d", i)),
		}
		p <- m
		fmt.Println("produce:", m)
	}

	go func() {
		wg.Done()
		cancel()
	}()

	// signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	select {
	case <-ctx.Done():
	case <-interrupt:
	}

	err = q.Stop(ctx)
	if err != nil {
		panic(err)
	}

	cancel()

}

// func TestNewQueue(t *testing.T) {
//	tracing.InitOnlyTracingLog("pushLeastMessages")
//	tracinglogger.SetLevel(tracinglogger.DebugLevel)
//
//	c := &queue.Config{}
//	c.Type = "memory"
//	q, err := NewQueue(c)
//	if err != nil {
//		panic(err)
//	}
//	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
//	defer cancel()
//
//	p := make(chan *queue.Message, 1024)
//	cc := make(chan *queue.Message, 1024)
//	ack := make(chan *queue.Message, 1024)
//	err = q.ProduceTopic(ctx, "pushLeastMessages", p)
//	if err != nil {
//		t.Fatal(err.Error())
//		return
//	}
//	err = q.ConsumeTopics(ctx, "pushLeastMessages", cc, ack)
//	if err != nil {
//		t.Fatal(err.Error())
//	}
//
//	for i := 0; i < 2; i++ {
//		err = q.ConsumeTopics(ctx, "pushLeastMessages", cc, ack)
//		if err != nil {
//			panic(err)
//		}
//
//	}
//
//	size := 100
//	wg := sync.WaitGroup{}
//	wg.Add(size)
//	go func() {
//		for {
//			select {
//			case m := <-cc:
//				fmt.Println(string(m.Row))
//				ack <- m
//				wg.Done()
//			case <-ctx.Done():
//				fmt.Println("done...")
//				return
//			}
//		}
//	}()
//
//	for i := 0; i < size; i++ {
//		m := &queue.Message{
//			ID:  fmt.Sprintf("%d", i),
//			Row: []byte(fmt.Sprintf("hello%d", i)),
//		}
//		p <- m
//		fmt.Println("produce:", m)
//	}
//
//	go func() {
//		wg.Done()
//		cancel()
//	}()
//
//	// signal
//	interrupt := make(chan os.Signal, 1)
//	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
//	defer signal.Stop(interrupt)
//
//	select {
//	case <-ctx.Done():
//	case <-interrupt:
//	}
//
//	ctxStop, can := context.WithTimeout(context.Background(), 10*time.Second)
//	defer can()
//	err = q.Stop(ctxStop)
//	if err != nil {
//		panic(err)
//	}
//
//	cancel()
// }

func Test_partition(t *testing.T) {
	type args struct {
		id  string
		num int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			"p1",
			args{id: "1", num: 3},
			1,
		},
		{
			"p2",
			args{id: "4", num: 3},
			1,
		},
		{
			"p0",
			args{id: "3", num: 3},
			2,
		},
		{
			"p3",
			args{id: "5", num: 3},
			2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := partition(tt.args.id, tt.args.num); got != tt.want {
				t.Errorf("partition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_marshalTracingMessage(t *testing.T) {
	m := &queue.Message{
		Row: []byte("hello"),
	}
	message, err := marshalTracingMessage(m, "test")
	if err != nil {
		panic(err)
	}
	tracingMessage, err := unmarshalTracingMessage(message, "test")
	if err != nil {
		panic(err)
	}
	log := logger.ContextLog(tracingMessage.Context)
	log.Printf("%#v\n", tracingMessage)
}
