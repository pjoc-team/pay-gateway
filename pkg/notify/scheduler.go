package notify

import (
	"github.com/blademainer/commons/pkg/recoverable"
	pay "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
	"time"
)

type Scheduler struct {
	queue         Queue              `json:"-" yaml:"-"`
	QueueConfig   *QueueConfig       `json:"queue_config" yaml:"QueueConfig"`
	NoticeCh      chan pay.PayNotice `json:"-" yaml:"-"`
	done          chan bool          `json:"-" yaml:"-"`
	stopped       bool               `json:"-" yaml:"-"`
	noticeService *NotifyService     `json:"-" yaml:"-"`
	Concurrency   int                `json:"concurrency" yaml:"Concurrency"`
}

func InitScheduler(config *QueueConfig, concurrency int, noticeService *NotifyService) (scheduler *Scheduler, err error) {
	log := logger.Log()

	queue, err := InstanceQueue(*config, noticeService)
	if err != nil {
		return
	}

	scheduler = &Scheduler{}
	scheduler.NoticeCh = make(chan pay.PayNotice, concurrency)
	scheduler.done = make(chan bool, 1)
	scheduler.queue = queue
	log.Infof("InitScheduler... queue: %v", queue)
	log.Infof("InitScheduler... scheduler.queue: %v", scheduler.queue)
	scheduler.noticeService = noticeService
	scheduler.stopped = false
	scheduler.Concurrency = concurrency
	scheduler.QueueConfig = config
	return
}

func (s *Scheduler) Start() {
	go s.startConsumer()
	go s.startNotice()
}

func (s *Scheduler) Stop() {
	s.stopped = true
	s.done <- true
}

func (s *Scheduler) startConsumer() {
	log := logger.Log()

	defer recoverable.Recover()
	for !s.stopped {
		notices, e := s.queue.Pull()
		if e != nil {
			log.Errorf("Failed to pull! error: %v", e)
			continue
		} else if notices == nil || len(notices) == 0 {
			time.Sleep(time.Second)
		} else {
			log.Infof("Pulled notices: %v", notices)
		}
		for _, notice := range notices {
			s.NoticeCh <- *notice
		}
	}
}

func (s *Scheduler) startNotice() {
	for i := 0; i < s.Concurrency; i++ {
		go s.startThreads()
	}
}

func (s *Scheduler) startThreads() {
	for !s.stopped {
		select {
		case notice := <-s.NoticeCh:
			s.notice(&notice)
		}
	}
}

func (s *Scheduler) notice(payNotice *pay.PayNotice) {
	log := logger.Log()

	defer recoverable.Recover()
	err := s.noticeService.SendPayNotice(payNotice)
	if err != nil {
		log.Errorf("Failed to send notice! order: %v error: %v", payNotice, err.Error())
		err = s.noticeService.UpdatePayNoticeFail(payNotice, err)
		if err != nil {
			log.Errorf("Failed to update db! notice: %v error: %v", payNotice, err.Error())
		} else {
			log.Debugf("Success to update notice! notice: %v", payNotice)
		}
		return
	}
	err = s.noticeService.UpdatePayNoticeSuccess(payNotice)
	if err != nil {
		log.Errorf("Failed to update notice ok! notice: %v error: %v", payNotice, err.Error())
	}

}
