package do

import "sync"

func Do(fn func()) <-chan struct{} {
	done := make(chan struct{}, 1)
	go func() {
		defer func() {
			done <- struct{}{}
			close(done)
		}()
		fn()
	}()
	return done
}

func Parallel(fns ...func()) {
	var wg sync.WaitGroup
	for _, fn := range fns {
		go WithGroup(&wg, fn)
	}
	wg.Wait()
}

func WithGroup(wg *sync.WaitGroup, fn func()) {
	wg.Add(1)
	defer wg.Done() // in case fn panics
	fn()
}

func WithLock(lk sync.Locker, fn func()) {
	lk.Lock()
	defer lk.Unlock() // in case fn panics
	fn()
}
