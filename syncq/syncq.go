package syncq

import (
	"container/list"
	"context"
)

// 效果类似于channel，但是可以设置为无数量限制(max<=0)
// Enqueue 接口会阻塞直到可以元素放入队列中
// Dequeue 接口会阻塞直到队列中有元素返回
type SyncQueue struct {
	ctx    context.Context
	cancel context.CancelFunc

	l   *list.List
	max int
	in  chan interface{} // use to enqueue
	out chan interface{} // use to dequeue
}

// max代表队列元素个数上限，若小于等于0，则队列无元素上限
// 内部会启动一个goroutine用于channel同步，可用Destroy()方法销毁。
// 注意调用Destroy()后就不可执行入队出队操作，否则会一直阻塞下去。
func NewSyncQueue(max int) *SyncQueue {
	ctx, cancel := context.WithCancel(context.Background())
	q := &SyncQueue{
		ctx:    ctx,
		cancel: cancel,
		l:      list.New(),
		max:    max,
		in:     make(chan interface{}),
		out:    make(chan interface{}),
	}
	go q.dispatch()
	return q
}

func (q *SyncQueue) dispatch() {
	for {
		if q.l.Len() == 0 {
			// 只允许入队列
			select {
			case v := <-q.in:
				q.l.PushBack(v)
			case <-q.ctx.Done():
				return
			}
		}
		e := q.l.Front()
		if q.max > 0 && q.l.Len() >= q.max {
			// 到达队列最大元素个数限制，只允许出队列
			select {
			case q.out <- e.Value:
				q.l.Remove(e)
			case <-q.ctx.Done():
				return
			}
		} else {
			// 允许入队列和出队列
			select {
			case value := <-q.in:
				q.l.PushBack(value)
			case q.out <- e.Value:
				q.l.Remove(e)
			case <-q.ctx.Done():
				return
			}
		}
	}
}

func (q *SyncQueue) Enqueue(value interface{}) {
	q.in <- value
}

func (q *SyncQueue) Dequeue() interface{} {
	return <-q.out
}

func (q *SyncQueue) Destroy() {
	// cancel dispatch goroutine
	q.cancel()
}
