package vcache_test

import (
	"testing"
	"time"

	"github.com/imajinyun/knifer-go/vcache"
)

type facadeTicker struct{}

func (facadeTicker) Stop() {}

func TestFacadeTimedCache(t *testing.T) {
	c := vcache.NewTimed[string, int](50 * time.Millisecond)
	c.Put("x", 10)
	if v, ok := c.Get("x"); !ok || v != 10 {
		t.Fatalf("expected x=10, got %v, ok=%v", v, ok)
	}
	time.Sleep(100 * time.Millisecond)
	if _, ok := c.Get("x"); ok {
		t.Fatal("expected 'x' to expire from timed cache")
	}
}

func TestFacadeNewTimedScheduled(t *testing.T) {
	c := vcache.NewTimedScheduled[string, int](time.Hour, 10*time.Minute)
	c.Put("a", 1)
	if v, ok := c.Get("a"); !ok || v != 1 {
		t.Fatalf("expected a=1, got %v, ok=%v", v, ok)
	}
	c.CancelPruneSchedule()
}

func TestFacadeNewWeak(t *testing.T) {
	v := 42
	c := vcache.NewWeak[string, int](time.Hour)
	c.Put("a", &v)
	if got, ok := c.Get("a"); !ok || *got != 42 {
		t.Fatalf("expected a=&42, got %v, ok=%v", got, ok)
	}
}

func TestFacadeCacheWithClock(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	now := base
	timed := vcache.NewTimedWithOptions[string, int](
		vcache.WithTimeout[string, int](time.Second),
		vcache.WithClock[string, int](func() time.Time { return now }),
	)
	timed.Put("a", 1)
	now = base.Add(2 * time.Second)
	if _, ok := timed.Get("a"); ok {
		t.Fatal("expected timed cache entry to expire with custom clock")
	}

	now = base
	weak := vcache.NewWeakWithOptions[string, int](
		vcache.WithTimeout[string, *int](time.Second),
		vcache.WithClock[string, *int](func() time.Time { return now }),
	)
	v := 7
	weak.Put("a", &v)
	now = base.Add(2 * time.Second)
	if _, ok := weak.Get("a"); ok {
		t.Fatal("expected weak cache entry to expire with custom clock")
	}
}

func TestFacadeCacheWithTickerFactory(t *testing.T) {
	called := false
	ticks := make(chan time.Time)
	timed := vcache.NewTimedWithOptions[string, int](
		vcache.WithTickerFactory[string, int](func(delay time.Duration) (<-chan time.Time, vcache.Ticker) {
			called = true
			if delay != time.Second {
				t.Fatalf("ticker delay = %s, want 1s", delay)
			}
			return ticks, facadeTicker{}
		}),
	)
	timed.SchedulePrune(time.Second)
	timed.CancelPruneSchedule()
	if !called {
		t.Fatal("facade ticker factory was not called")
	}
}
