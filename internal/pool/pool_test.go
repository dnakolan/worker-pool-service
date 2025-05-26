package pool

import (
	"context"
	"testing"
	"time"

	"github.com/dnakolan/worker-pool-service/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// waitForJobStatus waits for a job to reach a specific status with timeout
func waitForJobStatus(t *testing.T, pool *WorkerPool, jobID string, expectedStatus model.JobStatus) *model.Job {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if job, exists := pool.GetJob(context.Background(), jobID); exists {
			if job.Status == expectedStatus {
				return job
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("Job %s did not reach status %s within timeout", jobID, expectedStatus)
	return nil
}

// waitForNJobsWithStatus waits for n jobs to reach a specific status
func waitForNJobsWithStatus(t *testing.T, pool *WorkerPool, n int, status model.JobStatus) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		count := 0
		for _, job := range pool.GetAllJobs(context.Background(), &model.JobFilter{Status: &status}) {
			if job.Status == status {
				count++
			}
		}
		if count >= n {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("Did not reach %d jobs with status %s within timeout", n, status)
}

func TestWorkerPool_Basic(t *testing.T) {
	// Create a pool with 2 workers and queue size of 5
	ctx := context.Background()
	pool := NewWorkerPool(ctx, 2, 5)
	pool.Start()
	defer pool.Stop()

	// Create a test job
	job := &model.Job{
		UID:     uuid.New(),
		Type:    "sleep",
		Payload: model.SleepJobPayload{Duration: "100ms"},
		Status:  model.JobStatusPending,
	}

	// Submit the job
	err := pool.SubmitJob(ctx, job)
	assert.NoError(t, err)

	// Wait for job completion
	completedJob := waitForJobStatus(t, pool, job.UID.String(), model.JobStatusCompleted)
	assert.NotNil(t, completedJob.StartedAt)
	assert.NotNil(t, completedJob.CompletedAt)

	// Verify result
	result, ok := completedJob.Result.(model.SleepJobResult)
	assert.True(t, ok)
	assert.Equal(t, "100ms", result.SleptFor)
}

func TestWorkerPool_Concurrent(t *testing.T) {
	// Create a pool with 3 workers and queue size of 10
	ctx := context.Background()
	pool := NewWorkerPool(ctx, 3, 10)
	pool.Start()
	defer pool.Stop()

	// Create multiple jobs
	numJobs := 5
	jobs := make([]*model.Job, numJobs)
	for i := 0; i < numJobs; i++ {
		jobs[i] = &model.Job{
			UID:     uuid.New(),
			Type:    "math",
			Payload: model.MathJobPayload{Number: i + 1},
			Status:  model.JobStatusPending,
		}
	}

	// Submit jobs concurrently
	errCh := make(chan error, numJobs)
	for _, job := range jobs {
		go func(j *model.Job) {
			errCh <- pool.SubmitJob(ctx, j)
		}(job)
	}

	// Wait for all submissions
	for i := 0; i < numJobs; i++ {
		assert.NoError(t, <-errCh)
	}

	// Wait for all jobs to complete
	waitForNJobsWithStatus(t, pool, numJobs, model.JobStatusCompleted)

	// Check all jobs completed with correct results
	for _, job := range jobs {
		completedJob, exists := pool.GetJob(ctx, job.UID.String())
		assert.True(t, exists)
		assert.Equal(t, model.JobStatusCompleted, completedJob.Status)

		result, ok := completedJob.Result.(model.MathJobResult)
		assert.True(t, ok)

		// Verify math result
		payload := job.Payload.(model.MathJobPayload)
		expected := 0
		for i := 0; i < payload.Number; i++ {
			expected += i
		}
		assert.Equal(t, expected, result.Result)
	}
}

func TestWorkerPool_Cancellation(t *testing.T) {
	// Create a pool with context
	ctx, cancel := context.WithCancel(context.Background())
	pool := NewWorkerPool(ctx, 2, 5)
	pool.Start()
	defer pool.Stop()

	// Create a long-running job
	job := &model.Job{
		UID:     uuid.New(),
		Type:    "sleep",
		Payload: model.SleepJobPayload{Duration: "5s"},
		Status:  model.JobStatusPending,
	}

	// Submit the job
	err := pool.SubmitJob(ctx, job)
	assert.NoError(t, err)

	// Wait for job to start
	runningJob := waitForJobStatus(t, pool, job.UID.String(), model.JobStatusRunning)
	assert.NotNil(t, runningJob)

	// Cancel context
	cancel()

	// Wait for job to fail
	failedJob := waitForJobStatus(t, pool, job.UID.String(), model.JobStatusFailed)
	assert.Contains(t, failedJob.Error, "context canceled")
}

func TestWorkerPool_QueueFull(t *testing.T) {
	// Create a pool with small queue and NO workers
	ctx := context.Background()
	pool := NewWorkerPool(ctx, 0, 1) // Set workers to 0 so jobs stay in queue
	pool.Start()
	defer pool.Stop()

	// Create jobs that will fill the queue
	job1 := &model.Job{
		UID:     uuid.New(),
		Type:    "sleep",
		Payload: model.SleepJobPayload{Duration: "200ms"},
		Status:  model.JobStatusPending,
	}
	job2 := &model.Job{
		UID:     uuid.New(),
		Type:    "sleep",
		Payload: model.SleepJobPayload{Duration: "200ms"},
		Status:  model.JobStatusPending,
	}

	// Submit first job (should succeed)
	err := pool.SubmitJob(ctx, job1)
	assert.NoError(t, err)

	// Try to submit second job (should fail with queue full)
	err = pool.SubmitJob(ctx, job2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "job queue is full")

	// Verify job1 is still in pending state (since there are no workers)
	job, exists := pool.GetJob(ctx, job1.UID.String())
	assert.True(t, exists)
	assert.Equal(t, model.JobStatusPending, job.Status)
}

func TestWorkerPool_GetAllJobs(t *testing.T) {
	ctx := context.Background()
	pool := NewWorkerPool(ctx, 2, 5)
	pool.Start()
	defer pool.Stop()

	// Create jobs with different types and statuses
	jobs := []*model.Job{
		{
			UID:     uuid.New(),
			Type:    "sleep",
			Payload: model.SleepJobPayload{Duration: "100ms"},
			Status:  model.JobStatusPending,
		},
		{
			UID:     uuid.New(),
			Type:    "math",
			Payload: model.MathJobPayload{Number: 5},
			Status:  model.JobStatusPending,
		},
	}

	// Submit all jobs
	for _, job := range jobs {
		err := pool.SubmitJob(ctx, job)
		assert.NoError(t, err)
	}

	// Wait for jobs to complete
	waitForNJobsWithStatus(t, pool, len(jobs), model.JobStatusCompleted)

	// Test filtering by type
	sleepType := "sleep"
	mathType := "math"
	completedStatus := model.JobStatusCompleted

	// Filter by sleep type
	sleepJobs := pool.GetAllJobs(ctx, &model.JobFilter{
		Type: &sleepType,
	})
	assert.Len(t, sleepJobs, 1)
	assert.Equal(t, "sleep", sleepJobs[0].Type)

	// Filter by math type
	mathJobs := pool.GetAllJobs(ctx, &model.JobFilter{
		Type: &mathType,
	})
	assert.Len(t, mathJobs, 1)
	assert.Equal(t, "math", mathJobs[0].Type)

	// Filter by status
	completedJobs := pool.GetAllJobs(ctx, &model.JobFilter{
		Status: &completedStatus,
	})
	assert.Len(t, completedJobs, 2)
	for _, job := range completedJobs {
		assert.Equal(t, model.JobStatusCompleted, job.Status)
	}
}

func TestWorkerPool_InvalidJobType(t *testing.T) {
	ctx := context.Background()
	pool := NewWorkerPool(ctx, 1, 5)
	pool.Start()
	defer pool.Stop()

	// Create a job with invalid type
	job := &model.Job{
		UID:     uuid.New(),
		Type:    "invalid",
		Payload: model.SleepJobPayload{Duration: "100ms"},
		Status:  model.JobStatusPending,
	}

	// Submit the job
	err := pool.SubmitJob(ctx, job)
	assert.NoError(t, err)

	// Wait for job to fail
	failedJob := waitForJobStatus(t, pool, job.UID.String(), model.JobStatusFailed)
	assert.Contains(t, failedJob.Error, "unknown job type")
}

func TestExecuteJob(t *testing.T) {
	pool := &WorkerPool{}

	tests := []struct {
		name    string
		job     *model.Job
		want    model.JobResult
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid math job",
			job: &model.Job{
				Type:    "math",
				Payload: model.MathJobPayload{Number: 4},
			},
			want: model.MathJobResult{
				Result: 6, // sum of 0,1,2,3
			},
			wantErr: false,
		},
		{
			name: "invalid math payload type",
			job: &model.Job{
				Type:    "math",
				Payload: model.SleepJobPayload{Duration: "1s"},
			},
			wantErr: true,
			errMsg:  "invalid math payload type",
		},
		{
			name: "invalid job type",
			job: &model.Job{
				Type:    "invalid",
				Payload: model.MathJobPayload{Number: 1},
			},
			wantErr: true,
			errMsg:  "unknown job type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := pool.executeJob(tt.job)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestGetAllJobs_Filtering(t *testing.T) {
	pool := &WorkerPool{
		jobs: make(map[string]*model.Job),
	}

	// Create test jobs
	sleepJob := &model.Job{
		UID:    uuid.New(),
		Type:   "sleep",
		Status: model.JobStatusCompleted,
	}
	mathJob := &model.Job{
		UID:    uuid.New(),
		Type:   "math",
		Status: model.JobStatusPending,
	}

	// Store jobs
	pool.jobs[sleepJob.UID.String()] = sleepJob
	pool.jobs[mathJob.UID.String()] = mathJob

	// Test cases
	tests := []struct {
		name       string
		filter     *model.JobFilter
		wantTypes  []string
		wantStatus []model.JobStatus
		wantLen    int
	}{
		{
			name:    "no filter",
			filter:  &model.JobFilter{},
			wantLen: 2,
		},
		{
			name: "filter by type - sleep",
			filter: &model.JobFilter{
				Type: stringPtr("sleep"),
			},
			wantTypes: []string{"sleep"},
			wantLen:   1,
		},
		{
			name: "filter by status - completed",
			filter: &model.JobFilter{
				Status: jobStatusPtr(model.JobStatusCompleted),
			},
			wantStatus: []model.JobStatus{model.JobStatusCompleted},
			wantLen:    1,
		},
		{
			name: "filter by type and status - no match",
			filter: &model.JobFilter{
				Type:   stringPtr("sleep"),
				Status: jobStatusPtr(model.JobStatusPending),
			},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pool.GetAllJobs(nil, tt.filter)
			assert.Len(t, got, tt.wantLen)

			if tt.wantTypes != nil {
				for _, job := range got {
					assert.Contains(t, tt.wantTypes, job.Type)
				}
			}
			if tt.wantStatus != nil {
				for _, job := range got {
					assert.Contains(t, tt.wantStatus, job.Status)
				}
			}
		})
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func jobStatusPtr(s model.JobStatus) *model.JobStatus {
	return &s
}
