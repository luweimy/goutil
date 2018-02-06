package workerq

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/multierr"

	"github.com/luweimy/goutil/syncq"
)

type WorkerFunc func(worker *Worker) error

type Worker struct {
	ctx    context.Context
	cancel context.CancelFunc
	work   WorkerFunc
	errs   error
}

func NewWorker(ctx context.Context, work WorkerFunc) *Worker {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)
	return &Worker{
		ctx:    ctx,
		cancel: cancel,
		work:   work,
	}
}

func (c *Worker) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c *Worker) Wait() error {
	<-c.ctx.Done()
	return c.Err()
}

func (c *Worker) Err() error {
	return c.errs
}

func (c *Worker) Do() {
	defer func() {
		if err := recover(); err != nil {
			c.errs = multierr.Append(c.errs, fmt.Errorf("panic: %v", err))
		}
		c.cancel() // mark work done
	}()
	if c.work != nil {
		err := c.work(c)
		if err != nil {
			c.errs = multierr.Append(c.errs, err)
		}
	}
}

type WorkerQueue struct {
	workers chan *Worker // working workers
	mu      sync.RWMutex

	backlog *syncq.SyncQueue // backlog queue
}

func NewWorkerQueue(concurrency int) *WorkerQueue {
	if concurrency <= 0 {
		concurrency = 1
	}
	q := &WorkerQueue{
		workers: make(chan *Worker, concurrency),
		backlog: syncq.NewSyncQueue(),
	}
	go q.dispatchWorkers()
	return q
}

func (q *WorkerQueue) SetConcurrency(concurrency int) {
	if cap(q.workers) == concurrency {
		return
	}
	// resize workers channel, acquire wlock
	withLock(&q.mu, func() {
		if cap(q.workers) == concurrency {
			return
		}
		q.workers = make(chan *Worker, concurrency)
	})
}

func (q *WorkerQueue) AppendWorker(worker *Worker) <-chan struct{} {
	q.backlog.Enqueue(worker)
	return worker.Done()
}

func (q *WorkerQueue) AppendWorkerFunc(ctx context.Context, wf WorkerFunc) *Worker {
	worker := NewWorker(ctx, wf)
	q.AppendWorker(worker)
	return worker
}

func (q *WorkerQueue) dispatchWorkers() {
	for {
		worker := q.backlog.Dequeue().(*Worker)
		// working worker, acquire rlock
		go withLock(q.mu.RLocker(), func() {
			q.workers <- worker
			defer func() { <-q.workers }()
			worker.Do()
		})
	}
}

func withLock(lk sync.Locker, fn func()) {
	lk.Lock()
	defer lk.Unlock() // in case fn panics
	fn()
}
