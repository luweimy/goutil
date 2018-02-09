package async

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
