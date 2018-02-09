package do

import (
	"testing"
	"time"
)

func TestDo(t *testing.T) {
	t1 := time.Now()
	<-Do(func() {
		time.Sleep(time.Second)
		t.Log("await done")
	})

	t.Log("after await")
	t2 := time.Now()
	if t2.Sub(t1) < time.Second {
		t.Error("await do fn err")
	}
}
