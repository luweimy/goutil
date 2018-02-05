package loop

import (
	"sync"
	"time"
)

type CallFunc func(loop *CallLoop, now time.Time) bool

type CallLoop struct {
	once    sync.Once
	ticker  *time.Ticker
	handler CallFunc
	cancel  chan struct{}
}

func NewCallLoop(d time.Duration, handler CallFunc) *CallLoop {
	l := &CallLoop{
		ticker:  time.NewTicker(d),
		handler: handler,
		cancel:  make(chan struct{}, 1),
	}
	go l.loopCall()
	return l
}

func (l *CallLoop) loopCall() {
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

func (l *CallLoop) Wait() {
	<-l.cancel
}

func (l *CallLoop) Done() <-chan struct{} {
	return l.cancel
}

func (l *CallLoop) Stop() {
	l.once.Do(func() {
		l.ticker.Stop()
		l.cancel <- struct{}{}
		close(l.cancel)
	})
}
