package looper

import (
	"sync"
	"time"
)

type TimeLooper struct {
	once    sync.Once
	ticker  *time.Ticker
	handler func(time.Time) bool
	cancel  chan bool
}

func NewTimeLooper(d time.Duration, handler func(time.Time) bool) *TimeLooper {
	l := &TimeLooper{
		ticker:  time.NewTicker(d),
		handler: handler,
		cancel:  make(chan bool, 1),
	}
	go l.start()
	return l
}

func (l *TimeLooper) start() {
	for {
		select {
		case d := <-l.ticker.C:
			if l.handler != nil {
				if con := l.handler(d); !con {
					l.Stop()
					return
				}
			}
		case <-l.cancel:
			l.ticker.Stop()
			return
		}
	}
}

func (l *TimeLooper) Stop() {
	l.once.Do(func() {
		l.cancel <- true
		close(l.cancel)
	})
}
