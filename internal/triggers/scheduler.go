package triggers

import (
	"context"
	"fmt"
	"time"

	"github.com/logimos/conduktr/internal/engine"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// SchedulerTrigger implements cron-based scheduling
type SchedulerTrigger struct {
	cron   *cron.Cron
	engine *engine.Engine
	logger *zap.Logger
	jobs   map[string]cron.EntryID
}

// ScheduleConfig holds scheduling configuration
type ScheduleConfig struct {
	Jobs []JobConfig `yaml:"jobs"`
}

// JobConfig defines a scheduled job
type JobConfig struct {
	Name      string                 `yaml:"name"`
	Schedule  string                 `yaml:"schedule"` // Cron expression
	EventType string                 `yaml:"event_type"`
	Data      map[string]interface{} `yaml:"data"`
	Enabled   bool                   `yaml:"enabled"`
}

// NewSchedulerTrigger creates a new scheduler trigger
func NewSchedulerTrigger(config ScheduleConfig, engine *engine.Engine, logger *zap.Logger) *SchedulerTrigger {
	// Create cron scheduler with seconds precision
	c := cron.New(cron.WithSeconds())

	return &SchedulerTrigger{
		cron:   c,
		engine: engine,
		logger: logger,
		jobs:   make(map[string]cron.EntryID),
	}
}

// Start begins the scheduler
func (s *SchedulerTrigger) Start() error {
	s.logger.Info("Scheduler trigger started")
	s.cron.Start()
	return nil
}

// Stop stops the scheduler
func (s *SchedulerTrigger) Stop() error {
	s.cron.Stop()
	s.logger.Info("Scheduler trigger stopped")
	return nil
}

// AddJob adds a new scheduled job
func (s *SchedulerTrigger) AddJob(job JobConfig) error {
	if !job.Enabled {
		s.logger.Info("Job disabled, skipping", zap.String("name", job.Name))
		return nil
	}

	entryID, err := s.cron.AddFunc(job.Schedule, func() {
		s.executeJob(job)
	})

	if err != nil {
		return fmt.Errorf("failed to add cron job '%s': %w", job.Name, err)
	}

	s.jobs[job.Name] = entryID
	s.logger.Info("Scheduled job added",
		zap.String("name", job.Name),
		zap.String("schedule", job.Schedule),
		zap.String("event_type", job.EventType))

	return nil
}

// RemoveJob removes a scheduled job
func (s *SchedulerTrigger) RemoveJob(name string) error {
	entryID, exists := s.jobs[name]
	if !exists {
		return fmt.Errorf("job '%s' not found", name)
	}

	s.cron.Remove(entryID)
	delete(s.jobs, name)

	s.logger.Info("Scheduled job removed", zap.String("name", name))
	return nil
}

// executeJob executes a scheduled job
func (s *SchedulerTrigger) executeJob(job JobConfig) {
	s.logger.Info("Executing scheduled job",
		zap.String("name", job.Name),
		zap.String("event_type", job.EventType))

	// Create context with scheduler-specific metadata
	contextData := map[string]interface{}{
		"trigger_type": "scheduler",
		"job_name":     job.Name,
		"schedule":     job.Schedule,
		"event_type":   job.EventType,
		"timestamp":    time.Now().Unix(),
		"execution_id": fmt.Sprintf("sched_%d", time.Now().UnixNano()),
	}

	// Merge job data with context
	for k, v := range job.Data {
		contextData[k] = v
	}

	// Execute workflow asynchronously
	go executeWorkflow(context.Background(), s.engine, s.logger, job.EventType, contextData)
}

// ListJobs returns all scheduled jobs
func (s *SchedulerTrigger) ListJobs() []JobStatus {
	entries := s.cron.Entries()
	var jobs []JobStatus

	for name, entryID := range s.jobs {
		for _, entry := range entries {
			if entry.ID == entryID {
				jobs = append(jobs, JobStatus{
					Name:     name,
					Schedule: "", // Would need to store this separately
					Next:     entry.Next,
					Prev:     entry.Prev,
				})
				break
			}
		}
	}

	return jobs
}

// JobStatus represents the status of a scheduled job
type JobStatus struct {
	Name     string    `json:"name"`
	Schedule string    `json:"schedule"`
	Next     time.Time `json:"next"`
	Prev     time.Time `json:"prev"`
}

// GetJobStatus returns status of a specific job
func (s *SchedulerTrigger) GetJobStatus(name string) (JobStatus, error) {
	entryID, exists := s.jobs[name]
	if !exists {
		return JobStatus{}, fmt.Errorf("job '%s' not found", name)
	}

	entries := s.cron.Entries()
	for _, entry := range entries {
		if entry.ID == entryID {
			return JobStatus{
				Name: name,
				Next: entry.Next,
				Prev: entry.Prev,
			}, nil
		}
	}

	return JobStatus{}, fmt.Errorf("job entry not found")
}

// UpdateJob updates an existing job
func (s *SchedulerTrigger) UpdateJob(job JobConfig) error {
	// Remove existing job if it exists
	if _, exists := s.jobs[job.Name]; exists {
		if err := s.RemoveJob(job.Name); err != nil {
			return fmt.Errorf("failed to remove existing job: %w", err)
		}
	}

	// Add updated job
	return s.AddJob(job)
}
