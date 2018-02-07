package syncq

import (
	"container/list"
	"errors"
	"runtime"
	"testing"
	"time"
)

func TestSyncQueue(t *testing.T) {
	q := NewSyncQueue()
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

func TestSyncQueue_Destroy(t *testing.T) {
	numgo := runtime.NumGoroutine()
	q := NewSyncQueue()

	numgo2 := runtime.NumGoroutine()
	if numgo+1 != numgo2 {
		t.Errorf("except %v, actual %v", numgo+1, numgo2)
	}

	q.Enqueue(1)
	q.Enqueue(1)
	q.Destroy()

	// Delay一下，SyncQueue的goroutine才会退出
	time.Sleep(time.Millisecond)

	numgo3 := runtime.NumGoroutine()
	if numgo != numgo3 {
		t.Errorf("except %v, actual %v", numgo, numgo3)
	}
}

func TestNewSyncQueue(t *testing.T) {
	q := NewSyncQueueWithSize(2)
	q.Enqueue(1)
	q.Enqueue(2)

	var err = errors.New("enqueue not complete")
	go func() {
		q.Enqueue(2)
		err = nil // 此处必会运行
		q.Enqueue(2)
		t.Errorf("enqueue unable to complete") // 此处不会运行
	}()
	go q.Dequeue()
	time.Sleep(time.Millisecond * 500)
	if err != nil {
		t.Error(err)
	}
}

func BenchmarkSyncQueue(b *testing.B) {
	q := NewSyncQueue()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Enqueue(i)
		q.Dequeue()
	}
}

func BenchmarkList(b *testing.B) {
	q := list.New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.PushBack(i)
		q.Remove(q.Front())
	}
}
