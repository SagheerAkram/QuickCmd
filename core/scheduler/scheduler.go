package scheduler

import (
	"fmt"
	"time"
)

// Scheduler manages scheduled command execution
type Scheduler struct {
	jobs map[string]*ScheduledJob
}

// ScheduledJob represents a scheduled command
type ScheduledJob struct {
	ID          string
	Name        string
	Command     string
	Schedule    string // Cron expression
	NextRun     time.Time
	LastRun     time.Time
	Enabled     bool
	RunCount    int
	FailCount   int
	UserID      string
	Notifications []string
}

// NewScheduler creates a new scheduler
func NewScheduler() *Scheduler {
	return &Scheduler{
		jobs: make(map[string]*ScheduledJob),
	}
}

// Schedule schedules a new job
func (s *Scheduler) Schedule(name, command, cronExpr, userID string) (*ScheduledJob, error) {
	nextRun, err := parseNextRun(cronExpr)
	if err != nil {
		return nil, fmt.Errorf("invalid cron expression: %w", err)
	}
	
	job := &ScheduledJob{
		ID:       generateID(),
		Name:     name,
		Command:  command,
		Schedule: cronExpr,
		NextRun:  nextRun,
		Enabled:  true,
		UserID:   userID,
	}
	
	s.jobs[job.ID] = job
	return job, nil
}

// Run executes due jobs
func (s *Scheduler) Run() []*ScheduledJob {
	now := time.Now()
	executed := []*ScheduledJob{}
	
	for _, job := range s.jobs {
		if job.Enabled && job.NextRun.Before(now) {
			executed = append(executed, job)
			job.LastRun = now
			job.RunCount++
			job.NextRun, _ = parseNextRun(job.Schedule)
		}
	}
	
	return executed
}

func parseNextRun(cronExpr string) (time.Time, error) {
	// Simplified - would use proper cron parser
	return time.Now().Add(1 * time.Hour), nil
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
