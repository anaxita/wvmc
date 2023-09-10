package scheduler

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type Scheduler struct {
	lg   *zap.SugaredLogger
	jobs []Job
}

func NewScheduler(l *zap.SugaredLogger) *Scheduler {
	return &Scheduler{
		lg: l,
	}
}

type Job struct {
	Name       string
	Interval   time.Duration
	IsStartNow bool
	Func       JobFunc
}

type JobFunc func(ctx context.Context) error

// AddJob adds a job to the scheduler. Interval must be greater than 0.
func (s *Scheduler) AddJob(name string, interval time.Duration, isStartNow bool, fn JobFunc) *Scheduler {
	if interval <= 0 {
		s.lg.Panicw("interval must be greater than 0", "job_name", name)
	}

	job := Job{
		Name:       name,
		Interval:   interval,
		IsStartNow: isStartNow,
		Func:       fn,
	}

	s.jobs = append(s.jobs, job)

	return s
}

func (s *Scheduler) Start(ctx context.Context) {
	for _, job := range s.jobs {
		go s.startJob(ctx, job)
	}
}

func (s *Scheduler) startJob(ctx context.Context, job Job) {
	t := time.NewTimer(job.Interval)
	defer t.Stop()

	if job.IsStartNow {
		t.Reset(0)
	}

	for {
		select {
		case <-ctx.Done():
			s.lg.Info("scheduler stopped")
			return
		case <-t.C:
			s.lg.Infow("start job", "job_name", job.Name)

			err := job.Func(ctx)
			if err != nil {
				s.lg.Errorw("job failed", "job_name", job.Name, "err", err)
			}

			t.Reset(job.Interval)
		}
	}
}
