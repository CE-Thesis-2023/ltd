package custcron

import (
	"context"
	"time"

	"github.com/go-co-op/gocron"
)

type Options struct {
	register RegisterFunc
}

type Optioner func(o *Options)

type RegisterFunc func(s *gocron.Scheduler) error

func WithRegisterFunc(f RegisterFunc) Optioner {
	return func(o *Options) {
		o.register = f
	}
}

func New(options ...Optioner) *Scheduler {
	opts := &Options{}
	for _, opt := range options {
		opt(opts)
	}

	s := gocron.NewScheduler(time.UTC)

	sched := &Scheduler{
		s:       s,
		options: opts,
	}
	return sched
}

type Scheduler struct {
	s       *gocron.Scheduler
	options *Options
}

func (s *Scheduler) Start() error {
	reg := s.options.register

	if reg != nil {
		if err := reg(s.s); err != nil {
			return err
		}
	}

	s.s.StartBlocking()
	return nil
}

func (s *Scheduler) Stop(ctx context.Context) error {
	if s.s != nil {
		s.s.StopBlockingChan()
	}
	return nil
}

func (s *Scheduler) Name() string {
	return "primary_scheduler"
}
