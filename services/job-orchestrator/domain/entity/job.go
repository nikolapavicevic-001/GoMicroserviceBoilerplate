package entity

import (
	"encoding/json"
	"time"
)

// JobStatus represents the status of a job execution
type JobStatus string

const (
	JobStatusSubmitted JobStatus = "SUBMITTED"
	JobStatusRunning   JobStatus = "RUNNING"
	JobStatusFinished  JobStatus = "FINISHED"
	JobStatusFailed    JobStatus = "FAILED"
	JobStatusCancelled JobStatus = "CANCELLED"
)

// Job represents a Flink job definition
type Job struct {
	ID          string
	Name        string
	JarURL      string
	EntryClass  string
	Config      map[string]interface{}
	ScheduleCron *string
	Enabled     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Validate validates the job entity
func (j *Job) Validate() error {
	if j.Name == "" {
		return ErrInvalidJobName
	}
	if j.JarURL == "" {
		return ErrInvalidJarURL
	}
	if j.EntryClass == "" {
		return ErrInvalidEntryClass
	}
	return nil
}

// JobExecution represents a single execution of a job
type JobExecution struct {
	ID           string
	JobID        string
	FlinkJobID   *string
	FlinkJarID   *string
	Status       JobStatus
	StartedAt    *time.Time
	FinishedAt   *time.Time
	ErrorMessage *string
	CreatedAt    time.Time
}

// IsRunning returns true if the job execution is currently running
func (e *JobExecution) IsRunning() bool {
	return e.Status == JobStatusRunning || e.Status == JobStatusSubmitted
}

// ConfigJSON returns the config as JSON bytes
func (j *Job) ConfigJSON() ([]byte, error) {
	if j.Config == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(j.Config)
}

