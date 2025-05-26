package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSleepJobPayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		payload SleepJobPayload
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty duration",
			payload: SleepJobPayload{},
			wantErr: true,
			errMsg:  "duration is required",
		},
		{
			name: "valid duration",
			payload: SleepJobPayload{
				Duration: "1s",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payload.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestJob_UnmarshalJSON(t *testing.T) {
	testUUID := uuid.New()
	now := time.Now()

	tests := []struct {
		name    string
		json    string
		want    Job
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid sleep job",
			json: `{
				"uid": "` + testUUID.String() + `",
				"type": "sleep",
				"payload": {"duration": "1s"},
				"status": "pending",
				"created_at": "` + now.Format(time.RFC3339) + `"
			}`,
			want: Job{
				UID:       testUUID,
				Type:      "sleep",
				Payload:   SleepJobPayload{Duration: "1s"},
				Status:    JobStatusPending,
				CreatedAt: &now,
			},
			wantErr: false,
		},
		{
			name: "valid math job",
			json: `{
				"uid": "` + testUUID.String() + `",
				"type": "math",
				"payload": {"number": 42},
				"status": "pending",
				"created_at": "` + now.Format(time.RFC3339) + `"
			}`,
			want: Job{
				UID:       testUUID,
				Type:      "math",
				Payload:   MathJobPayload{Number: 42},
				Status:    JobStatusPending,
				CreatedAt: &now,
			},
			wantErr: false,
		},
		{
			name: "invalid job type",
			json: `{
				"uid": "` + testUUID.String() + `",
				"type": "invalid",
				"payload": {},
				"status": "pending",
				"created_at": "` + now.Format(time.RFC3339) + `"
			}`,
			wantErr: true,
			errMsg:  "unknown job type: invalid",
		},
		{
			name: "invalid sleep payload",
			json: `{
				"uid": "` + testUUID.String() + `",
				"type": "sleep",
				"payload": {"invalid": "field"},
				"status": "pending",
				"created_at": "` + now.Format(time.RFC3339) + `"
			}`,
			wantErr: true,
			errMsg:  "invalid sleep job payload",
		},
		{
			name: "invalid math payload",
			json: `{
				"uid": "` + testUUID.String() + `",
				"type": "math",
				"payload": {"number": "not a number"},
				"status": "pending",
				"created_at": "` + now.Format(time.RFC3339) + `"
			}`,
			wantErr: true,
			errMsg:  "invalid math job payload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var job Job
			err := json.Unmarshal([]byte(tt.json), &job)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.UID, job.UID)
				assert.Equal(t, tt.want.Type, job.Type)
				assert.Equal(t, tt.want.Status, job.Status)
				assert.Equal(t, tt.want.Payload, job.Payload)
				assert.Equal(t, tt.want.CreatedAt.Format(time.RFC3339), job.CreatedAt.Format(time.RFC3339))
			}
		})
	}
}

func TestCreateJobRequest_ParsePayload(t *testing.T) {
	tests := []struct {
		name    string
		request CreateJobRequest
		want    JobPayload
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid sleep job",
			request: CreateJobRequest{
				Type:    "sleep",
				Payload: json.RawMessage(`{"duration": "1s"}`),
			},
			want:    SleepJobPayload{Duration: "1s"},
			wantErr: false,
		},
		{
			name: "valid math job",
			request: CreateJobRequest{
				Type:    "math",
				Payload: json.RawMessage(`{"number": 42}`),
			},
			want:    MathJobPayload{Number: 42},
			wantErr: false,
		},
		{
			name: "invalid job type",
			request: CreateJobRequest{
				Type:    "invalid",
				Payload: json.RawMessage(`{}`),
			},
			wantErr: true,
			errMsg:  "type is invalid",
		},
		{
			name: "invalid sleep payload",
			request: CreateJobRequest{
				Type:    "sleep",
				Payload: json.RawMessage(`{}`),
			},
			wantErr: true,
			errMsg:  "duration is required",
		},
		{
			name: "invalid math payload format",
			request: CreateJobRequest{
				Type:    "math",
				Payload: json.RawMessage(`{"number": "not a number"}`),
			},
			wantErr: true,
			errMsg:  "invalid math job payload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.request.ParsePayload()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestJobStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{"pending status", "pending", true},
		{"running status", "running", true},
		{"completed status", "completed", true},
		{"failed status", "failed", true},
		{"invalid status", "invalid", false},
		{"empty status", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test IsValidJobStatus
			assert.Equal(t, tt.want, IsValidJobStatus(tt.status))

			// Test ParseJobStatus
			status, err := ParseJobStatus(tt.status)
			if tt.want {
				assert.NoError(t, err)
				assert.Equal(t, JobStatus(tt.status), status)
			} else {
				assert.Error(t, err)
				assert.Equal(t, JobStatus(""), status)
			}
		})
	}
}

func TestJobPayloadType(t *testing.T) {
	tests := []struct {
		name    string
		payload JobPayload
		want    string
	}{
		{
			name:    "sleep payload",
			payload: SleepJobPayload{Duration: "1s"},
			want:    "sleep",
		},
		{
			name:    "math payload",
			payload: MathJobPayload{Number: 42},
			want:    "math",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.payload.Type())
		})
	}
}

func TestJobResult(t *testing.T) {
	tests := []struct {
		name   string
		result JobResult
		want   string
	}{
		{
			name:   "sleep result",
			result: SleepJobResult{SleptFor: "1s"},
			want:   "sleep",
		},
		{
			name:   "math result",
			result: MathJobResult{Result: 42},
			want:   "math",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.result.Type())
		})
	}
}
