package model

import "fmt"

type JobFilter struct {
	Type   *string    `json:"type,omitempty"`
	Status *JobStatus `json:"status,omitempty"`
}

func (f *JobFilter) Validate() error {
	if f == nil {
		return nil
	}

	if f.Type != nil {
		if *f.Type != "math" && *f.Type != "sleep" {
			return fmt.Errorf("unsupported job type")
		}
	}

	if f.Status != nil {
		if *f.Status == "" {
			return fmt.Errorf("status cannot be empty")
		}
		if !IsValidJobStatus(string(*f.Status)) {
			return fmt.Errorf("invalid status: %s", *f.Status)
		}
	}

	return nil
}
