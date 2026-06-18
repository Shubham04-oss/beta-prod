package unified

import (
	"context"
	"log"
	"sync"
)

type Job struct {
	TenantID  string
	OrgID     string
	ProductID string
	Action    string // "UPSERT" or "DELETE"
}

type WorkerPool struct {
	jobs    chan Job
	wg      sync.WaitGroup
	service *Service
	workers int
}

func NewWorkerPool(service *Service, numWorkers int, queueSize int) *WorkerPool {
	return &WorkerPool{
		jobs:    make(chan Job, queueSize),
		service: service,
		workers: numWorkers,
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i)
	}
}

func (wp *WorkerPool) worker(ctx context.Context, id int) {
	defer wp.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-wp.jobs:
			if !ok {
				return
			}
			log.Printf("[Worker %d] Processing Job for Product: %s (Action: %s)", id, job.ProductID, job.Action)
			err := wp.service.ProcessPush(ctx, job.TenantID, job.OrgID, job.ProductID, job.Action)
			if err != nil {
				log.Printf("[Worker %d] Failed to push product %s: %v", id, job.ProductID, err)
			}
		}
	}
}

func (wp *WorkerPool) Enqueue(job Job) {
	wp.jobs <- job
}

func (wp *WorkerPool) Stop() {
	close(wp.jobs)
	wp.wg.Wait()
}
