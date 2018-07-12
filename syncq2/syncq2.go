package syncq2

import (
	"container/list"
	"sync"
)

// SyncQueue相当于无容量限制的channel
// Enqueue 接口将元素放入队列中，不会发生阻塞
// Dequeue 接口会阻塞直到队列中有元素返回，阻塞的情况只在队列空的时候才会出现
type SyncQueue struct {
	cond *sync.Cond
	l    *list.List

	in      chan interface{}
	out     chan interface{}
	inOnce  sync.Once
	outOnce sync.Once
}

func New() *SyncQueue {
	q := &SyncQueue{
		l:    list.New(),
		cond: sync.NewCond(&sync.Mutex{}),
	}
	return q
}

func (q *SyncQueue) Enqueue(value interface{}) {
	withLock(q.cond.L, func() {
		q.l.PushBack(value)
		q.cond.Signal()
	})
}

func (q *SyncQueue) Dequeue() interface{} {
	var v interface{}
	withLock(q.cond.L, func() {
		// if queue is empty, wait enqueue
		for q.l.Len() <= 0 {
			q.cond.Wait()
		}
		v = q.l.Remove(q.l.Front())
	})
	return v
}

func (q *SyncQueue) EnqueueC() chan<- interface{} {
	if q.in == nil {
		q.inOnce.Do(func() {
			q.in = make(chan interface{})
			go func() {
				q.Enqueue(<-q.in)
			}()
		})
	}
	return q.in
}

func (q *SyncQueue) DequeueC() <-chan interface{} {
	if q.out == nil {
		q.outOnce.Do(func() {
			q.out = make(chan interface{})
			go func() {
				q.out <- q.Dequeue()
			}()
		})
	}
	return q.out
}

func withLock(lk sync.Locker, fn func()) {
	lk.Lock()
	defer lk.Unlock() // in case fn panics
	fn()
}
