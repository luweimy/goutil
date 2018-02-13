package signal

import (
	"os"
	"os/signal"
	"syscall"
)

func Listen(fn func(os.Signal), sig ...os.Signal) {
	termCh := make(chan os.Signal)
	signal.Notify(termCh, sig...)
	s := <-termCh
	signal.Stop(termCh)
	switch s {
	case syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2:
		fn(s)
	}
}

func TerminalListen(fn func(os.Signal)) {
	termCh := make(chan os.Signal)
	signal.Notify(termCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)
	sig := <-termCh
	signal.Stop(termCh)
	switch sig {
	case syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2:
		fn(sig)
	}
}
