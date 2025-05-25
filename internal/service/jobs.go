package service

import (
	"context"

	"github.com/dnakolan/worker-pool-service/internal/model"
	"github.com/dnakolan/worker-pool-service/internal/pool"
	"github.com/google/uuid"
)

type JobsService interface {
	CreateJobs(ctx context.Context, req *model.Job) error
	ListJobs(ctx context.Context, filter *model.JobFilter) ([]*model.Job, error)
	GetJobs(ctx context.Context, uid uuid.UUID) (*model.Job, error)
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

func (s *jobsService) GetJobs(ctx context.Context, uid uuid.UUID) (*model.Job, error) {
	// TODO: Implement get job by id
	return nil, nil
}
