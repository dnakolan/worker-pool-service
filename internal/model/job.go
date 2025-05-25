package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

type Job struct {
	UID         uuid.UUID  `json:"uid"`
	Type        string     `json:"type"`
	Payload     JobPayload `json:"payload"`
	Status      JobStatus  `json:"status"`
	Result      JobResult  `json:"result,omitempty"`
	Error       string     `json:"error,omitempty"`
	CreatedAt   *time.Time `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// JobPayload is an interface that all job payloads must implement
type JobPayload interface {
	Type() string
	Validate() error
}

// SleepJobPayload represents the payload for a sleep job
type SleepJobPayload struct {
	Duration string `json:"duration"`
}

func (p SleepJobPayload) Type() string {
	return "sleep"
}

func (p SleepJobPayload) Validate() error {
	if p.Duration == "" {
		return errors.New("duration is required")
	}
	return nil
}

// MathJobPayload represents the payload for a math job
type MathJobPayload struct {
	Number int `json:"number"`
}

func (p MathJobPayload) Type() string {
	return "math"
}

func (p MathJobPayload) Validate() error {
	return nil
}

// UnmarshalJSON implements custom JSON unmarshaling for Job
func (j *Job) UnmarshalJSON(data []byte) error {
	// First unmarshal into a temporary struct with a generic payload
	type tempJob struct {
		UID         uuid.UUID       `json:"uid"`
		Type        string          `json:"type"`
		Payload     json.RawMessage `json:"payload"`
		Status      JobStatus       `json:"status"`
		Result      json.RawMessage `json:"result,omitempty"`
		Error       string          `json:"error,omitempty"`
		CreatedAt   time.Time       `json:"created_at"`
		StartedAt   time.Time       `json:"started_at,omitempty"`
		CompletedAt time.Time       `json:"completed_at,omitempty"`
	}

	var temp tempJob
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Copy the simple fields
	j.UID = temp.UID
	j.Type = temp.Type
	j.Status = temp.Status
	j.Error = temp.Error
	j.CreatedAt = &temp.CreatedAt
	j.StartedAt = &temp.StartedAt
	j.CompletedAt = &temp.CompletedAt

	// Unmarshal the payload based on the job type
	switch temp.Type {
	case "sleep":
		var payload SleepJobPayload
		if err := json.Unmarshal(temp.Payload, &payload); err != nil {
			return fmt.Errorf("invalid sleep job payload: %w", err)
		}
		j.Payload = payload
	case "math":
		var payload MathJobPayload
		if err := json.Unmarshal(temp.Payload, &payload); err != nil {
			return fmt.Errorf("invalid math job payload: %w", err)
		}
		j.Payload = payload
	default:
		return fmt.Errorf("unknown job type: %s", temp.Type)
	}

	return nil
}

type JobResult interface {
	Type() string
}

type SleepJobResult struct {
	SleptFor string `json:"slept_for"`
}

func (r SleepJobResult) Type() string {
	return "sleep"
}

type MathJobResult struct {
	Result int `json:"result"`
}

func (r MathJobResult) Type() string {
	return "math"
}

type CreateJobRequest struct {
	Type    string          `json:"type" validate:"required"`
	Payload json.RawMessage `json:"payload"`
}

// ParsePayload validates the request and returns the appropriate JobPayload
func (r *CreateJobRequest) ParsePayload() (JobPayload, error) {
	if r.Type != "sleep" && r.Type != "math" {
		return nil, errors.New("type is invalid")
	}

	switch r.Type {
	case "sleep":
		var payload SleepJobPayload
		if err := json.Unmarshal(r.Payload, &payload); err != nil {
			return nil, fmt.Errorf("invalid sleep job payload: %w", err)
		}
		if err := payload.Validate(); err != nil {
			return nil, err
		}
		return payload, nil
	case "math":
		var payload MathJobPayload
		if err := json.Unmarshal(r.Payload, &payload); err != nil {
			return nil, fmt.Errorf("invalid math job payload: %w", err)
		}
		if err := payload.Validate(); err != nil {
			return nil, err
		}
		return payload, nil
	default:
		return nil, fmt.Errorf("unknown job type: %s", r.Type)
	}
}

// IsValidJobStatus checks if a string is a valid job status
func IsValidJobStatus(s string) bool {
	switch JobStatus(s) {
	case JobStatusPending, JobStatusRunning, JobStatusCompleted, JobStatusFailed:
		return true
	default:
		return false
	}
}

// ParseJobStatus converts a string to JobStatus, returning an error if invalid
func ParseJobStatus(s string) (JobStatus, error) {
	if !IsValidJobStatus(s) {
		return "", fmt.Errorf("invalid job status: %s", s)
	}
	return JobStatus(s), nil
}
