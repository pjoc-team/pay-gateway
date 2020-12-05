package notify

import (
	"context"
	"github.com/blademainer/commons/pkg/recoverable"
	pay "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
	"time"
)

// Scheduler scheduler
type Scheduler struct {
	ctx         context.Context
	QueueConfig *QueueConfig       `json:"queue_config" yaml:"QueueConfig"`
	NotifyCh    chan pay.PayNotify `json:"-" yaml:"-"`
	Concurrency int                `json:"concurrency" yaml:"Concurrency"`

	queue         Queue
	done          chan bool
	stopped       bool
	notifyService *Service
}

// InitScheduler init scheduler
func InitScheduler(
	ctx context.Context, config *QueueConfig, concurrency int,
	notifyService *Service,
) (scheduler *Scheduler, err error) {
	log := logger.Log()

	queue, err := InstanceQueue(*config, notifyService)
	if err != nil {
		return
	}

	scheduler = &Scheduler{}
	scheduler.ctx = ctx
	scheduler.NotifyCh = make(chan pay.PayNotify, concurrency)
	scheduler.done = make(chan bool, 1)
	scheduler.queue = queue
	log.Infof("InitScheduler... queue: %v", queue)
	log.Infof("InitScheduler... scheduler.queue: %v", scheduler.queue)
	scheduler.notifyService = notifyService
	scheduler.stopped = false
	scheduler.Concurrency = concurrency
	scheduler.QueueConfig = config
	return
}

// Start start server
func (s *Scheduler) Start(ctx context.Context) {
	go s.startConsumer(ctx)
	go s.startNotify(ctx)
}

// Stop stop server
func (s *Scheduler) Stop() {
	s.stopped = true
	s.done <- true
}

func (s *Scheduler) startConsumer(ctx context.Context) {
	log := logger.Log()

	defer recoverable.Recover()
	for !s.stopped {
		notifys, e := s.queue.Pull(ctx)
		if e != nil {
			log.Errorf("Failed to pull! error: %v", e)
			continue
		} else if len(notifys) == 0 {
			time.Sleep(time.Second)
			continue
		} else {
			log.Infof("Pulled notifys: %v", notifys)
		}
		for _, notify := range notifys {
			s.NotifyCh <- *notify
		}
	}
}

func (s *Scheduler) startNotify(ctx context.Context) {
	for i := 0; i < s.Concurrency; i++ {
		go s.startThreads(ctx)
	}
}

func (s *Scheduler) startThreads(ctx context.Context) {
	log := logger.ContextLog(ctx)
	for !s.stopped {
		select {
		case <-ctx.Done():
			log.Warn("scheduler's threads stopped!")
			s.stopped = true
			return
		case notify := <-s.NotifyCh:
			s.notify(ctx, &notify)
		}
	}
}

func (s *Scheduler) notify(ctx context.Context, payNotify *pay.PayNotify) {
	log := logger.Log()

	defer recoverable.Recover()
	err := s.notifyService.SendPayNotify(ctx, payNotify)
	if err != nil {
		log.Errorf("Failed to send notify! order: %v error: %v", payNotify, err.Error())
		err = s.notifyService.UpdatePayNotifyFail(ctx, payNotify, err)
		if err != nil {
			log.Errorf("Failed to update db! notify: %v error: %v", payNotify, err.Error())
		} else {
			log.Debugf("Success to update notify! notify: %v", payNotify)
		}
		return
	}
	err = s.notifyService.UpdatePayNotifySuccess(ctx, payNotify)
	if err != nil {
		log.Errorf("Failed to update notify ok! notify: %v error: %v", payNotify, err.Error())
	}

}
