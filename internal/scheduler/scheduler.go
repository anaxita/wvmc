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

func (s *Scheduler) AddJob(name string, interval time.Duration, isStartNow bool, fn JobFunc) *Scheduler {
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
	t := time.NewTicker(job.Interval)
	defer t.Stop()

	if job.IsStartNow {
		err := job.Func(ctx)
		if err != nil {
			s.lg.Errorw("job failed", "job_name", job.Name, zap.Error(err))
		}
	}

	for range t.C {
		err := job.Func(ctx)
		if err != nil {
			s.lg.Errorw("job failed", "job_name", job.Name, zap.Error(err))
		}
	}
}
