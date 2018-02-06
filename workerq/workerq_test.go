package workerq

import (
	"context"
	"sync"
	"testing"
	"time"
)

var TimeFormatLayout = "15:04:05.9999"

func TestNewWorker(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*1))
	w := NewWorker(ctx, func(worker *Worker) error {
		return nil
	})

	go cancel()
	w.Wait()
}

func TestNewWorkerQueue(t *testing.T) {
	wq := NewWorkerQueue(1)

	doneWorkers := make([]*Worker, 0, 2)
	mu := sync.Mutex{}

	// 启动第一个worker
	worker1 := wq.AppendWorkerFunc(nil, func(worker *Worker) error {
		t.Logf("worker1 start %p", worker)
		time.Sleep(time.Second)
		withLock(&mu, func() {
			doneWorkers = append(doneWorkers, worker)
		})
		t.Logf("worker1 done %p", worker)
		return nil
	})

	// 启动第二个worker
	worker2 := wq.AppendWorkerFunc(nil, func(worker *Worker) error {
		t.Logf("worker2 start %p", worker)
		withLock(&mu, func() {
			doneWorkers = append(doneWorkers, worker)
		})
		time.Sleep(time.Second)
		t.Logf("worker2 done %p", worker)
		return nil
	})

	// 等待两个worker处理完
	worker1.Wait()
	worker2.Wait()

	// 验证worker处理顺序
	withLock(&mu, func() {
		if len(doneWorkers) != 2 {
			t.Errorf("doneWorkers length except 2 %d", len(doneWorkers))
		}
		if doneWorkers[0] != worker1 {
			t.Errorf("doneWorkers except worker1 %p != %p", doneWorkers[0], worker1)
		}
		if doneWorkers[1] != worker2 {
			t.Errorf("doneWorkers except worker2 %p != %p", doneWorkers[1], worker2)
		}
	})
}

func TestNewWorkerQueueSetConcurrency(t *testing.T) {
	wq := NewWorkerQueue(1)

	// 并发1时，启动一个worker
	wq.AppendWorkerFunc(nil, func(worker *Worker) error {
		t.Log("worker0 start", time.Now().Format(TimeFormatLayout))
		time.Sleep(time.Millisecond * 500)
		t.Log("worker0 done", time.Now().Format(TimeFormatLayout))
		return nil
	})

	// 改为并发2
	wq.SetConcurrency(2)
	t.Log("cap(q.workers)", cap(wq.workers))

	startWorkers := make([]time.Time, 0, 3)
	mu := sync.Mutex{}

	worker1 := wq.AppendWorkerFunc(nil, func(worker *Worker) error {
		t.Log("worker1 start", time.Now().Format(TimeFormatLayout))
		withLock(&mu, func() {
			startWorkers = append(startWorkers, time.Now())
		})
		time.Sleep(time.Millisecond * 800)
		t.Log("worker1 done", time.Now().Format(TimeFormatLayout))
		return nil
	})

	worker2 := wq.AppendWorkerFunc(nil, func(worker *Worker) error {
		t.Log("worker2 start", time.Now().Format(TimeFormatLayout))
		withLock(&mu, func() {
			startWorkers = append(startWorkers, time.Now())
		})
		time.Sleep(time.Millisecond * 100)
		t.Log("worker2 done", time.Now().Format(TimeFormatLayout))
		return nil
	})

	worker1.Wait()
	worker2.Wait()

	withLock(&mu, func() {
		if len(startWorkers) != 2 {
			t.Errorf("startWorkers length except 2 %d", len(startWorkers))
		}
		distance := startWorkers[1].Sub(startWorkers[0])
		if distance > time.Millisecond*10 {
			t.Errorf("startWorkers time not except, %v", startWorkers)
		}
	})
}
