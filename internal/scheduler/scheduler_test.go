package scheduler

import (
	"context"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/jobs"
)

func TestSchedulerRegistersCronSpecAndRunsJobs(t *testing.T) {
	ran := 0
	s := New()
	if err := s.Register("job-1", "0 * * * *", func(context.Context) (jobs.Result, error) {
		ran++
		return jobs.Result{}, nil
	}); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := s.Start(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s.RunOnce(ctx)

	if ran != 1 {
		t.Fatalf("expected job to run once, got %d", ran)
	}
}

func TestSchedulerRejectsInvalidCronSpec(t *testing.T) {
	s := New()

	err := s.Register("job-1", "not-a-cron-spec", func(context.Context) (jobs.Result, error) {
		return jobs.Result{}, nil
	})
	if err == nil {
		t.Fatal("expected invalid cron spec error")
	}
}
