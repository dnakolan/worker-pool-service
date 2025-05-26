package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJobFilter_Validate(t *testing.T) {
	tests := []struct {
		name      string
		JobFilter *JobFilter
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "nil JobFilter",
			JobFilter: nil,
			wantErr:   false,
		},
		{
			name:      "empty JobFilter",
			JobFilter: &JobFilter{},
			wantErr:   false,
		},
		{
			name: "valid status",
			JobFilter: &JobFilter{
				Status: jobStatusPtr(JobStatusRunning),
			},
			wantErr: false,
		},
		{
			name: "invalid status",
			JobFilter: &JobFilter{
				Status: jobStatusPtr(JobStatus("bad")),
			},
			wantErr: true,
			errMsg:  "invalid status: bad",
		},
		{
			name: "empty status",
			JobFilter: &JobFilter{
				Status: jobStatusPtr(JobStatus("")),
			},
			wantErr: true,
			errMsg:  "status cannot be empty",
		},
		{
			name: "valid type",
			JobFilter: &JobFilter{
				Type: stringPtr("math"),
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			JobFilter: &JobFilter{
				Type: stringPtr("bad"),
			},
			wantErr: true,
			errMsg:  "unsupported job type",
		},
		{
			name: "empty type",
			JobFilter: &JobFilter{
				Type: stringPtr(""),
			},
			wantErr: true,
			errMsg:  "unsupported job type",
		},
		{
			name: "all valid parameters",
			JobFilter: &JobFilter{
				Type:   stringPtr("math"),
				Status: jobStatusPtr(JobStatusRunning),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.JobFilter.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Equal(t, tt.errMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func stringPtr(v string) *string {
	return &v
}

func jobStatusPtr(v JobStatus) *JobStatus {
	return &v
}
