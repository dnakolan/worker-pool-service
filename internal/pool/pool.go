package pool

import (
	"context"
	"sync"

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
		jobQueue:    make(chan *model.Job),
		resultQueue: make(chan *model.Job),
		quit:        make(chan struct{}),
		jobs:        make(map[string]*model.Job),
		numWorkers:  numWorkers,
		wg:          sync.WaitGroup{},
		ctx:         ctx,
		cancel:      cancel,
	}
}
