package cron

import (
	"sync"
	"testing"
)

func TestSchedulerExecutorAndRunnerConcurrentReplacement(t *testing.T) {
	s := NewSchedulerWithOptions(WithExecutor(func(fn func()) { fn() }), WithRunner(func(fn func()) { fn() }))
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			s.SetExecutor(func(fn func()) { fn() })
			s.SetRunner(func(fn func()) { fn() })
		}()
		go func() {
			defer wg.Done()
			s.submit(func() {})
			s.run(func() {})
		}()
	}
	wg.Wait()
}
