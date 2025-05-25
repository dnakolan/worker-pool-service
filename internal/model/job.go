package model

import (
	"errors"
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
	CreatedAt   time.Time  `json:"created_at"`
	StartedAt   time.Time  `json:"started_at,omitempty"`
	CompletedAt time.Time  `json:"completed_at,omitempty"`
}

type JobPayload interface{}

type SleepJobPayload struct {
	JobPayload
	Duration string `json:"duration"`
}

type MathJobPayload struct {
	JobPayload
	Number int `json:"number"`
}

type JobResult interface{}

type SleepJobResult struct {
	JobResult
	Duration string `json:"duration"`
}

type MathJobResult struct {
	JobResult
}

type CreateJobRequest struct {
	Type    string     `json:"type" validate:"required"`
	Payload JobPayload `json:"payload"`
}

func (r *CreateJobRequest) Validate() error {
	if r.Type != "sleep" && r.Type != "math" {
		return errors.New("type is invalid")
	}
	return nil
}
