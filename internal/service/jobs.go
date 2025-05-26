package service

import (
	"context"
	"errors"

	"github.com/dnakolan/worker-pool-service/internal/model"
	"github.com/dnakolan/worker-pool-service/internal/pool"
)

type JobsService interface {
	CreateJobs(ctx context.Context, req *model.Job) error
	ListJobs(ctx context.Context, filter *model.JobFilter) ([]*model.Job, error)
	GetJobs(ctx context.Context, uid string) (*model.Job, error)
}

type jobsService struct {
	pool *pool.WorkerPool
}

func NewJobsService(pool *pool.WorkerPool) *jobsService {
	return &jobsService{pool: pool}
}

func (s *jobsService) CreateJobs(ctx context.Context, req *model.Job) error {
	return s.pool.SubmitJob(ctx, req)
}

func (s *jobsService) ListJobs(ctx context.Context, filter *model.JobFilter) ([]*model.Job, error) {
	jobs := s.pool.GetAllJobs(ctx, filter)

	if jobs == nil {
		return make([]*model.Job, 0), nil
	}
	return jobs, nil
}

func (s *jobsService) GetJobs(ctx context.Context, uid string) (*model.Job, error) {
	job, exists := s.pool.GetJob(ctx, uid)
	if !exists {
		return nil, errors.New("job not found")
	}
	return job, nil
}
