package worker

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

// Pool is a bounded worker pool that processes Jobs concurrently.
type Pool struct {
	jobs      chan Job
	wg        sync.WaitGroup
	processor *Processor
	log       *zap.Logger
	size      int
}

// NewPool creates a Pool with the given number of workers and job queue capacity.
func NewPool(poolSize, queueSize int, processor *Processor, log *zap.Logger) *Pool {
	return &Pool{
		jobs:      make(chan Job, queueSize),
		processor: processor,
		log:       log,
		size:      poolSize,
	}
}

// Run starts worker goroutines that consume jobs until ctx is cancelled
// or the jobs channel is closed.
func (p *Pool) Run(ctx context.Context) {
	p.log.Info("worker pool starting", zap.Int("workers", p.size))

	for i := 0; i < p.size; i++ {
		p.wg.Add(1)
		go p.runWorker(ctx, i)
	}

	p.wg.Wait()
	p.log.Info("worker pool stopped")
}

// Enqueue sends a job to the pool's job channel.
// Blocks if the channel is full. Returns false if ctx is cancelled before
// the job could be enqueued.
func (p *Pool) Enqueue(ctx context.Context, job Job) bool {
	select {
	case p.jobs <- job:
		p.log.Debug("job enqueued", zap.String("file", job.FilePath))
		return true
	case <-ctx.Done():
		p.log.Warn("enqueue cancelled, context done", zap.String("file", job.FilePath))
		return false
	}
}

// Shutdown closes the job channel and waits for all in-flight jobs to complete.
// Must be called after the producer (scanner) has stopped sending jobs.
func (p *Pool) Shutdown() {
	close(p.jobs)
	p.wg.Wait()
}

// runWorker is the main loop for a single worker goroutine.
func (p *Pool) runWorker(ctx context.Context, id int) {
	defer p.wg.Done()

	log := p.log.With(zap.Int("worker_id", id))
	log.Info("worker started")

	for {
		select {
		case job, ok := <-p.jobs:
			if !ok {
				log.Info("worker stopped: job channel closed")
				return
			}
			p.processJob(ctx, log, job)

		case <-ctx.Done():
			log.Info("worker draining remaining jobs")
			for job := range p.jobs {
				p.processJob(ctx, log, job)
			}
			log.Info("worker stopped: context cancelled")
			return
		}
	}
}

// processJob executes a single job and logs the outcome.
func (p *Pool) processJob(ctx context.Context, log *zap.Logger, job Job) {
	log.Info("processing job", zap.String("file", job.FilePath))

	if err := p.processor.Process(ctx, job); err != nil {
		log.Error("job failed",
			zap.String("file", job.FilePath),
			zap.Error(err),
		)
		return
	}

	log.Info("job completed", zap.String("file", job.FilePath))
}
