package call

import (
	"sync"
	"time"
)

type CallFunc func(loop *loopCall, now time.Time) bool

type loopCall struct {
	once    sync.Once
	ticker  *time.Ticker
	handler CallFunc
	cancel  chan struct{}
}

func Loop(interval time.Duration, handler CallFunc) *loopCall {
	l := &loopCall{
		ticker:  time.NewTicker(interval),
		handler: handler,
		cancel:  make(chan struct{}, 1),
	}
	go l.loopCall()
	return l
}

func (l *loopCall) loopCall() {
	for {
		select {
		case d := <-l.ticker.C:
			if con := l.handler(l, d); !con {
				l.Stop()
				return
			}
		}
	}
}

func (l *loopCall) Wait() {
	<-l.cancel
}

func (l *loopCall) Done() <-chan struct{} {
	return l.cancel
}

func (l *loopCall) Stop() {
	l.once.Do(func() {
		l.ticker.Stop()
		l.cancel <- struct{}{}
		close(l.cancel)
	})
}
