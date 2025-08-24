package shutdown

import (
	"log/slog"
	"sync"
	"time"
)

var sh = &shutdownmanager{
	mu:           sync.Mutex{},
	isShutdown:   false,
	shudownFuncs: make([]func(), 0),
}

func AddFunc(fn func()) {
	sh.AddFunc(fn)
}

func Shutdown(wait time.Duration) {
	sh.Shutdown(wait)
}

type shutdownmanager struct {
	mu           sync.Mutex
	isShutdown   bool
	shudownFuncs []func()
}

func (s *shutdownmanager) AddFunc(fn func()) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isShutdown {
		slog.Error("attempted to add shutdown func after shutdown")
		return
	}

	s.shudownFuncs = append(s.shudownFuncs, fn)
}

func (s *shutdownmanager) Shutdown(wait time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isShutdown {
		slog.Error("double shutdown attempted")
		return
	}

	if len(s.shudownFuncs) == 0 {
		return
	}

	var wg sync.WaitGroup
	for _, v := range s.shudownFuncs {
		wg.Add(1)
		fn := v
		go func() {
			fn()
			wg.Done()
		}()
	}

	ch := make(chan int, 1)
	go func() {
		wg.Wait()
		ch <- 0
	}()

	select {
	case <-time.After(30 * time.Second):
		slog.Warn("failed to shutdown gracefully")
		return
	case <-ch:
		slog.Info("completed graceful shutdown")
		return
	}
}
