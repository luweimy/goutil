package workerq

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/multierr"

	"github.com/luweimy/goutil/syncq2"
)

type WorkerFunc func(worker *Worker) error

type Worker struct {
	ctx    context.Context
	cancel context.CancelFunc
	begin  chan struct{} // use to notify worker process begin
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
		begin:  make(chan struct{}, 1),
		work:   work,
	}
}

func (c *Worker) Begin() <-chan struct{} {
	return c.begin
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
		c.cancel() // notify work done
	}()

	c.begin <- struct{}{} // notify work begin
	if c.work != nil {
		err := c.work(c)
		if err != nil {
			c.errs = multierr.Append(c.errs, err)
		}
	}
}

type WorkerQueue struct {
	cancel context.CancelFunc
	ctx    context.Context

	backlog *syncq2.SyncQueue // backlog workers queue(unlimited size)
	workers chan *Worker      // processing workers
	mu      sync.RWMutex
}

// New create WorkerQueue object, max concurrency workers is allowed
func New(concurrency int) *WorkerQueue {
	if concurrency <= 0 {
		concurrency = 1
	}
	ctx, cancel := context.WithCancel(context.Background())
	q := &WorkerQueue{
		cancel:  cancel,
		ctx:     ctx,
		workers: make(chan *Worker, concurrency),
		backlog: syncq2.New(),
	}
	return q
}

// Start start process backlog workers.
func (q *WorkerQueue) Start() *WorkerQueue {
	go q.dispatchWorkers()
	return q
}

// Stop stop process backlog workers
// stop can not stop processing workers and the workers not in backlog will be processed.
func (q *WorkerQueue) Stop() {
	q.cancel()
}

// SetConcurrency set concurrency will be blocked until processing workers all done.
func (q *WorkerQueue) SetConcurrency(concurrency int) {
	if cap(q.workers) == concurrency {
		return
	}
	// resize workers channel, acquire wlock to block worker processing.
	// the workers channel must be empty when can hold wlock.
	withLock(&q.mu, func() {
		if cap(q.workers) == concurrency {
			return
		}
		q.workers = make(chan *Worker, concurrency)
	})
}

// NumProcessingWorkers return num of processing workers, top limit is concurrency.
func (q *WorkerQueue) NumProcessingWorkers() int {
	var n = 0
	withLock(q.mu.RLocker(), func() {
		n = len(q.workers)
	})
	return n
}

func (q *WorkerQueue) AddWorker(worker *Worker) <-chan struct{} {
	q.backlog.Enqueue(worker)
	return worker.Done()
}

func (q *WorkerQueue) AddWorkerFunc(ctx context.Context, wf WorkerFunc) *Worker {
	worker := NewWorker(ctx, wf)
	q.AddWorker(worker)
	return worker
}

func (q *WorkerQueue) dispatchWorkers() {
	for {
		select {
		case worker := <-q.backlog.DequeueC():
			q.dispatchWorker(worker.(*Worker))
		case <-q.ctx.Done():
			// Stop is called.
			return
		}
	}
}

func (q *WorkerQueue) dispatchWorker(worker *Worker) {
	// acquire rlock during work processing to block resize workers channel buffer.
	q.mu.RLock()
	// worker must be processed, even through ctx canceled.
	q.workers <- worker
	// avoid work processing block dispatch goroutine.
	go func() {
		defer func() { // in case worker do panics
			<-q.workers
			q.mu.RUnlock()
		}()
		worker.Do()
	}()
}

func withLock(lk sync.Locker, fn func()) {
	lk.Lock()
	defer lk.Unlock() // in case fn panics
	fn()
}
