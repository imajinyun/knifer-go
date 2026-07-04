package cache

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestGetOrLoadSingleFlightPerKey(t *testing.T) {
	c := NewLRU[string, int](16)
	start := make(chan struct{})
	var supplierCalls atomic.Int32

	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			got, err := c.GetOrLoad("shared", func() (int, error) {
				supplierCalls.Add(1)
				time.Sleep(time.Millisecond)
				return 42, nil
			})
			if err != nil {
				t.Errorf("GetOrLoad error = %v", err)
				return
			}
			if got != 42 {
				t.Errorf("GetOrLoad = %d, want 42", got)
			}
		}()
	}
	close(start)
	wg.Wait()
	if got := supplierCalls.Load(); got != 1 {
		t.Fatalf("supplier calls = %d, want 1", got)
	}
}

func TestGetOrLoadDifferentKeysDoNotShareLock(t *testing.T) {
	c := NewLRU[int, int](128)
	start := make(chan struct{})
	var supplierCalls atomic.Int32

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			got, err := c.GetOrLoad(i, func() (int, error) {
				supplierCalls.Add(1)
				return i, nil
			})
			if err != nil {
				t.Errorf("GetOrLoad(%d) error = %v", i, err)
				return
			}
			if got != i {
				t.Errorf("GetOrLoad(%d) = %d, want %d", i, got, i)
			}
		}()
	}
	close(start)
	wg.Wait()
	if got := supplierCalls.Load(); got != 32 {
		t.Fatalf("supplier calls = %d, want 32", got)
	}
}

func TestListenerReentryConcurrentWithMutations(t *testing.T) {
	c := NewLRU[int, int](4)
	var removed atomic.Int32
	var reentered atomic.Bool
	c.SetListener(CacheListenerFunc[int, int](func(key, value int) {
		removed.Add(1)
		_ = c.Size()
		_ = c.ContainsKey(key)
		if reentered.CompareAndSwap(false, true) {
			c.PutWithTimeout(1000+key, value, time.Millisecond)
		}
	}))

	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Put(i, i)
			_, _ = c.Get(i)
			c.Remove(i)
		}()
	}
	wg.Wait()
	if removed.Load() == 0 {
		t.Fatal("listener did not observe removals")
	}
}

func TestTimedCacheSchedulePruneConcurrentCancel(t *testing.T) {
	var runnerCalls atomic.Int32
	c := NewTimedWithOptions[string, int](
		WithTickerFactory[string, int](func(time.Duration) (<-chan time.Time, Ticker) {
			ticks := make(chan time.Time)
			ticker := &testTicker{stopped: make(chan struct{}, 1)}
			return ticks, ticker
		}),
		WithRunner[string, int](func(fn func()) {
			runnerCalls.Add(1)
			go fn()
		}),
	)

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			c.SchedulePrune(time.Second)
		}()
		go func() {
			defer wg.Done()
			c.CancelPruneSchedule()
		}()
	}
	wg.Wait()
	c.CancelPruneSchedule()
	if got := runnerCalls.Load(); got > 32 {
		t.Fatalf("runner calls = %d, want no more than attempted schedules", got)
	}
}
