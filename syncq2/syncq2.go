package syncq2

import (
	"container/list"
	"sync"

	"github.com/luweimy/goutil/call"
)

// SyncQueue相当于无容量限制的channel
// Enqueue 接口将元素放入队列中，不会发生阻塞
// Dequeue 接口会阻塞直到队列中有元素返回，阻塞的情况只在队列空的时候才会出现
type SyncQueue struct {
	cond *sync.Cond
	l    *list.List
	out  chan interface{} // use to dequeue
}

func New() *SyncQueue {
	q := &SyncQueue{
		l:    list.New(),
		cond: sync.NewCond(&sync.Mutex{}),
	}
	return q
}

func (q *SyncQueue) Enqueue(value interface{}) {
	call.WithLock(q.cond.L, func() {
		q.l.PushBack(value)
		q.cond.Signal()
	})
}

func (q *SyncQueue) Dequeue() interface{} {
	var v interface{}
	call.WithLock(q.cond.L, func() {
		// if queue is empty, wait enqueue
		for q.l.Len() <= 0 {
			q.cond.Wait()
		}
		v = q.l.Remove(q.l.Front())
	})
	return v
}

func (q *SyncQueue) EnqueueC() chan<- interface{} {
	var in = make(chan interface{})
	go func() {
		q.Enqueue(<-in)
	}()
	return in
}

func (q *SyncQueue) DequeueC() <-chan interface{} {
	var out = make(chan interface{})
	go func() {
		out <- q.Dequeue()
	}()
	return out
}
