package pool

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/dnakolan/worker-pool-service/internal/model"
)

type WorkerPool struct {
	// Channels
	jobQueue    chan *model.Job
	resultQueue chan *model.Job
	quit        chan struct{}

	// State management
	jobs      map[string]*model.Job
	jobsMutex sync.RWMutex

	// Pool configuration
	numWorkers int
	wg         sync.WaitGroup

	// Context
	ctx    context.Context
	cancel context.CancelFunc
}

func NewWorkerPool(ctx context.Context, numWorkers int, poolSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(ctx)

	return &WorkerPool{
		jobQueue:    make(chan *model.Job, poolSize),
		resultQueue: make(chan *model.Job, poolSize),
		quit:        make(chan struct{}),
		jobs:        make(map[string]*model.Job),
		numWorkers:  numWorkers,
		wg:          sync.WaitGroup{},
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (p *WorkerPool) SubmitJob(ctx context.Context, job *model.Job) error {
	select {
	case p.jobQueue <- job:
		p.storeJob(job)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-p.ctx.Done():
		return p.ctx.Err()
	default:
		return errors.New("job queue is full")
	}
}

func (p *WorkerPool) GetJob(ctx context.Context, id string) (*model.Job, bool) {
	p.jobsMutex.RLock()
	defer p.jobsMutex.RUnlock()
	job, exists := p.jobs[id]
	return job, exists
}

func (p *WorkerPool) GetAllJobs(ctx context.Context, filter *model.JobFilter) []*model.Job {
	p.jobsMutex.RLock()
	defer p.jobsMutex.RUnlock()
	jobs := make([]*model.Job, 0)
	for _, v := range p.jobs {
		if filter.Type != nil && *filter.Type != v.Type {
			continue
		}
		if filter.Status != nil && *filter.Status != v.Status {
			continue
		}
		jobs = append(jobs, v)
	}
	return jobs
}

func (p *WorkerPool) Start() {
	slog.Info("Starting worker pool", "workers", p.numWorkers)

	// Start workers
	for i := 0; i < p.numWorkers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}

	// Start result processor
	p.wg.Add(1)
	go p.resultProcessor()
}

func (p *WorkerPool) Stop() {
	slog.Info("Stopping worker pool")
	p.cancel()
	close(p.quit)
	p.wg.Wait()
	close(p.jobQueue)
	close(p.resultQueue)
}

// Core worker goroutine
func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()

	for {
		select {
		case job := <-p.jobQueue:
			p.processJob(id, job)
		case <-p.quit:
			slog.Info("Worker shutting down", "worker_id", id)
			return
		case <-p.ctx.Done():
			slog.Info("Worker context cancelled", "worker_id", id)
			return
		}
	}
}

func (p *WorkerPool) processJob(workerID int, job *model.Job) {
	slog.Info("Processing job", "worker_id", workerID, "job_id", job.UID)

	// Update job status
	now := time.Now()
	job.Status = model.JobStatusRunning
	job.StartedAt = &now
	p.storeJob(job)

	// Execute the job
	result, err := p.executeJob(job)

	// Update final status
	completedAt := time.Now()
	job.CompletedAt = &completedAt

	if err != nil {
		job.Status = model.JobStatusFailed
		job.Error = err.Error()
	} else {
		job.Status = model.JobStatusCompleted
		job.Result = result
	}

	// Send to result processor
	select {
	case p.resultQueue <- job:
	case <-p.ctx.Done():
		return
	}
}

func (p *WorkerPool) executeJob(job *model.Job) (model.JobResult, error) {
	switch job.Type {
	case "sleep":
		payload, ok := job.Payload.(model.SleepJobPayload)
		if !ok {
			return nil, errors.New("invalid sleep payload type")
		}

		duration, err := time.ParseDuration(payload.Duration)
		if err != nil {
			return nil, fmt.Errorf("invalid duration: %w", err)
		}

		select {
		case <-time.After(duration):
			return model.SleepJobResult{
				SleptFor: duration.String(),
			}, nil
		case <-p.ctx.Done():
			return nil, p.ctx.Err()
		}

	case "math":
		payload, ok := job.Payload.(model.MathJobPayload)
		if !ok {
			return nil, errors.New("invalid math payload type")
		}

		result := 0
		for i := 0; i < payload.Number; i++ {
			result += i
		}
		return model.MathJobResult{
			Result: result,
		}, nil

	default:
		return nil, errors.New("unknown job type")
	}
}

func (p *WorkerPool) resultProcessor() {
	defer p.wg.Done()

	for {
		select {
		case job := <-p.resultQueue:
			p.storeJob(job)
			slog.Info("Job completed", "job_id", job.UID, "status", job.Status)
		case <-p.quit:
			return
		case <-p.ctx.Done():
			return
		}
	}
}

func (p *WorkerPool) storeJob(job *model.Job) {
	p.jobsMutex.Lock()
	defer p.jobsMutex.Unlock()
	p.jobs[job.UID.String()] = job
}
