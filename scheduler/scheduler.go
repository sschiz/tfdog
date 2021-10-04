package scheduler

import (
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/google/uuid"
)

// Scheduler is mechanism to schedule a tasks.
type Scheduler struct {
	s        *gocron.Scheduler
	interval time.Duration

	mu   sync.Mutex
	jobs map[uuid.UUID]*gocron.Job
}

// NewScheduler returns new Scheduler instance.
// Interval determines after what time to start the jobs.
func NewScheduler(interval time.Duration) *Scheduler {
	return &Scheduler{
		interval: interval,
		s:        gocron.NewScheduler(time.UTC),
		jobs:     make(map[uuid.UUID]*gocron.Job),
	}
}

// Schedule schedules new job.
func (s *Scheduler) Schedule(job interface{}, params ...interface{}) (uuid.UUID, error) {
	j, err := s.s.Every(s.interval).Do(job, params...)
	if err != nil {
		return uuid.Nil, err
	}

	id := uuid.New()
	s.mu.Lock()
	s.jobs[id] = j
	s.mu.Unlock()

	return id, nil
}

// Start starts scheduler.
// Starting is non-blocking.
func (s *Scheduler) Start() {
	s.s.StartAsync()
}

// Stop stops scheduler.
func (s *Scheduler) Stop() {
	s.s.Stop()
}

// RemoveJob removes a job.
func (s *Scheduler) RemoveJob(id uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobs[id]
	if ok {
		delete(s.jobs, id)
		s.s.RemoveByReference(job)
	}
}
