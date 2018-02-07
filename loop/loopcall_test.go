package loop

import (
	"fmt"
	"testing"
	"time"
)

func TestNewCallLoop(t *testing.T) {
	loopCall := NewLoopCall(time.Second, func(loop *LoopCall, now time.Time) bool {
		fmt.Println("NOW=>", now.Format(time.RFC3339))
		loop.Stop()
		return true
	})

	go func() {
		time.Sleep(time.Second * 1)
		loopCall.Stop()
		loopCall.Stop()
	}()

	// wait for loop call done
	<-loopCall.Done()

	// already stop
	loopCall.Wait()
}

func BenchmarkCallLoop(b *testing.B) {
	for i := 0; i < b.N; i++ {
		call := NewLoopCall(time.Second, func(loop *LoopCall, now time.Time) bool {
			return false
		})
		call.Stop()
	}
}

func BenchmarkTicker(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ticker := time.NewTicker(time.Second)
		ticker.Stop()
	}
}
