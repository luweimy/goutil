package syncq2

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestSyncQueue(t *testing.T) {
	q := New()
	for i := 0; i < 1000; i++ {
		q.Enqueue(i)
	}
	for i := 0; i < 1000; i++ {
		v := q.Dequeue()
		if v != i {
			t.Errorf("except %v actual %v", i, v)
		}
	}

	var (
		exceptValue = 1003
	)
	go func() {
		if value := q.Dequeue(); value != exceptValue {
			t.Errorf("except %v, actual %v", exceptValue, value)
		}
	}()

	time.Sleep(time.Millisecond * 500)
	q.Enqueue(exceptValue)
}

func TestSyncQueue2(t *testing.T) {
	q := New()
	go func() {
		time.Sleep(time.Second)
		q.Enqueue(1)
	}()
	q.Dequeue()
}

func TestNewSyncQueue(t *testing.T) {
	q := New()
	q.Enqueue(1)
	q.Enqueue(2)

	var err = errors.New("enqueue not complete")
	go func() {
		q.Enqueue(2)
		err = nil // 此处必会运行
		q.Enqueue(2)
	}()
	go q.Dequeue()
	time.Sleep(time.Millisecond * 500)
	if err != nil {
		t.Error(err)
	}
}

func TestNewSyncQueue2(t *testing.T) {
	q := New()
	go func() {
		q.Enqueue(1)
	}()
	v := q.Dequeue()
	q.Destroy()
	t.Log("dequeue=>", v)
}

func TestSyncQueueDequeueC(t *testing.T) {
	q := New()
	in := q.EnqueueC()
	out := q.DequeueC()
	in <- 1
	if elem := <-out; elem != 1 {
		t.Error("queue channel err")
	}
}

func BenchmarkSyncQueue(b *testing.B) {
	q := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Enqueue(i)
		q.Dequeue()
	}
}

func BenchmarkSyncQueue2(b *testing.B) {
	q := New()
	b.ResetTimer()

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		for i := 0; i < b.N; i++ {
			q.Dequeue()
		}
		wg.Done()
	}()
	go func() {
		for i := 0; i < b.N; i++ {
			q.Enqueue(i)
		}
		wg.Done()
	}()

	wg.Wait()
}
