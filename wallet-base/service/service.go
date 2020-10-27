package service

import (
	"fmt"
	"time"

	"upex-wallet/wallet-base/util"

	"upex-wallet/wallet-base/newbitx/misclib/log"
)

type Service struct {
	worker   Worker
	interval time.Duration

	lastUpdateTime time.Time
	closeCh        chan struct{}
	closeFinishCh  chan struct{}
}

func New(w Worker) *Service {
	return NewWithInterval(w, time.Second*5)
}

func NewWithInterval(w Worker, interval time.Duration) *Service {
	return &Service{
		worker:        w,
		interval:      interval,
		closeCh:       make(chan struct{}),
		closeFinishCh: make(chan struct{}),
	}
}

func (s *Service) Start() error {
	err := s.worker.Init()
	if err != nil {
		log.Errorf("worker (%s) init failed, %v", s.worker.Name(), err)
		return err
	}

	log.Infof("worker (%s) started", s.worker.Name())

L:
	for {
		select {
		case <-s.closeCh:
			s.worker.Destroy()
			log.Infof("worker (%s) stopped", s.worker.Name())
			break L
		default:
			if time.Now().Sub(s.lastUpdateTime) >= s.interval {
				s.lastUpdateTime = time.Now()
				s.work()
			}
		}

		time.Sleep(time.Millisecond * 10)
	}

	close(s.closeFinishCh)
	return nil
}

func (s *Service) Stop() {
	close(s.closeCh)
	<-s.closeFinishCh
}

func (s *Service) work() {
	tag := fmt.Sprintf("worker (%s)", s.worker.Name())
	util.WithRecover(tag, s.worker.Work, nil)
}
