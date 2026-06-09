package scheduler

import (
	"context"

	"github.com/awesome-academy/golang_baoan_thao/internal/jobs"
	"github.com/robfig/cron/v3"
)

type Runner func(context.Context) (jobs.Result, error)

type entry struct {
	name string
	spec string
	run  Runner
}

type Scheduler struct {
	entries []entry
}

func New() *Scheduler {
	return &Scheduler{}
}

func (s *Scheduler) Register(name, spec string, run Runner) error {
	if _, err := cron.ParseStandard(spec); err != nil {
		return err
	}
	s.entries = append(s.entries, entry{name: name, spec: spec, run: run})
	return nil
}

func (s *Scheduler) Start(ctx context.Context) error {
	c := cron.New()
	for _, item := range s.entries {
		run := item.run
		if _, err := c.AddFunc(item.spec, func() {
			_, _ = run(ctx)
		}); err != nil {
			return err
		}
	}
	c.Start()
	go func() {
		<-ctx.Done()
		stopCtx := c.Stop()
		<-stopCtx.Done()
	}()
	return nil
}

func (s *Scheduler) RunOnce(ctx context.Context) {
	for _, item := range s.entries {
		_, _ = item.run(ctx)
	}
}
