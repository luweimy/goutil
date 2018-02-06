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
