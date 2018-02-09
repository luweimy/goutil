package loopcall

import (
	"sync"
	"time"
)

type CallFunc func(loop *LoopCall, now time.Time) bool

type LoopCall struct {
	once    sync.Once
	ticker  *time.Ticker
	handler CallFunc
	cancel  chan struct{}
}

func New(d time.Duration, handler CallFunc) *LoopCall {
	l := &LoopCall{
		ticker:  time.NewTicker(d),
		handler: handler,
		cancel:  make(chan struct{}, 1),
	}
	go l.loopCall()
	return l
}

func (l *LoopCall) loopCall() {
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

func (l *LoopCall) Wait() {
	<-l.cancel
}

func (l *LoopCall) Done() <-chan struct{} {
	return l.cancel
}

func (l *LoopCall) Stop() {
	l.once.Do(func() {
		l.ticker.Stop()
		l.cancel <- struct{}{}
		close(l.cancel)
	})
}
