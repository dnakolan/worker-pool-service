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

	return nil
}
