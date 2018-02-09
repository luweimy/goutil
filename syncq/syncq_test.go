package syncq

import (
	"errors"
	"runtime"
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

func TestSyncQueue_Destroy(t *testing.T) {
	numgo := runtime.NumGoroutine()
	q := New()

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
	q := NewWithSize(2)
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
	q := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Enqueue(i)
		q.Dequeue()
	}
}

func BenchmarkChannelBuffer1(b *testing.B) {
	ch := make(chan struct{}, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch <- struct{}{}
		<-ch
	}
}

func BenchmarkChannelNoBufferNoSelect(b *testing.B) {
	ch := make(chan struct{}, 0)
	b.ResetTimer()
	go func() {
		for {
			<-ch
		}
	}()
	for i := 0; i < b.N; i++ {
		ch <- struct{}{}
	}
}

func BenchmarkChannelNoBufferWithSelect(b *testing.B) {
	ch := make(chan struct{}, 0)
	b.ResetTimer()
	go func() {
		for {
			select {
			case <-ch:
			}
		}
	}()
	for i := 0; i < b.N; i++ {
		ch <- struct{}{}
	}
}
