package cache

import (
	"runtime"
	"testing"
	"time"
)

func TestApplyOptionsEmpty(t *testing.T) {
	cfg := applyOptions[string, int](nil)
	if cfg.capacity != 0 || cfg.timeout != 0 || cfg.listener != nil || cfg.clock != nil {
		t.Fatalf("expected zero config: %+v", cfg)
	}
}

func TestApplyOptionsWithCapacityAndTimeout(t *testing.T) {
	cfg := applyOptions[string, int]([]Option[string, int]{
		WithCapacity[string, int](100),
		WithTimeout[string, int](time.Minute),
	})
	if cfg.capacity != 100 {
		t.Fatalf("capacity = %d, want 100", cfg.capacity)
	}
	if cfg.timeout != time.Minute {
		t.Fatalf("timeout = %v, want 1m", cfg.timeout)
	}
}

func TestApplyOptionsWithListenerAndClock(t *testing.T) {
	listener := CacheListenerFunc[string, int](func(string, int) {})
	cfg := applyOptions[string, int]([]Option[string, int]{
		WithListener[string, int](listener),
		WithClock[string, int](time.Now),
	})
	if cfg.listener == nil {
		t.Fatal("listener should be set")
	}
	if cfg.clock == nil {
		t.Fatal("clock should be set")
	}
}

func TestWithTickerFactoryAndRunner(t *testing.T) {
	factory := TickerFactory(func(time.Duration) (<-chan time.Time, Ticker) { return nil, nil })
	runner := func(fn func()) { go fn() }
	cfg := applyOptions[string, int]([]Option[string, int]{
		WithTickerFactory[string, int](factory),
		WithRunner[string, int](runner),
	})
	if cfg.tickerFactory == nil {
		t.Fatal("tickerFactory should be set")
	}
	if cfg.runner == nil {
		t.Fatal("runner should be set")
	}
}

func TestNilProviderOptionsDoNotOverwriteConfiguredProviders(t *testing.T) {
	listener := CacheListenerFunc[string, int](func(string, int) {})
	clock := func() time.Time { return time.Unix(1, 0) }
	factory := TickerFactory(func(time.Duration) (<-chan time.Time, Ticker) { return nil, nil })
	runner := func(fn func()) { fn() }
	cfg := applyOptions[string, int]([]Option[string, int]{
		WithListener[string, int](listener),
		WithListener[string, int](nil),
		WithClock[string, int](clock),
		WithClock[string, int](nil),
		WithTickerFactory[string, int](factory),
		WithTickerFactory[string, int](nil),
		WithRunner[string, int](runner),
		WithRunner[string, int](nil),
	})
	if cfg.listener == nil || cfg.clock == nil || cfg.tickerFactory == nil || cfg.runner == nil {
		t.Fatalf("nil provider option overwrote configured provider: %+v", cfg)
	}
}

func TestWithWeakFinalizerFunc(t *testing.T) {
	type obj struct{}
	finalizer := func(v *obj, fn func(*obj)) {
		runtime.SetFinalizer(v, fn)
	}
	cfg := applyOptions[string, *obj]([]Option[string, *obj]{
		WithWeakFinalizerFunc[string, obj](finalizer),
	})
	if cfg.finalizer == nil {
		t.Fatal("finalizer should be set")
	}

	// nil option should not overwrite
	cfg2 := applyOptions[string, *obj]([]Option[string, *obj]{
		WithWeakFinalizerFunc[string, obj](nil),
	})
	if cfg2.finalizer != nil {
		t.Fatal("nil WithWeakFinalizerFunc should not set finalizer")
	}
}

func TestWithWeakFinalizerEnabled(t *testing.T) {
	type obj struct{}

	cfg := applyOptions[string, *obj]([]Option[string, *obj]{
		WithWeakFinalizerEnabled[string, obj](false),
	})
	if !cfg.finalizerOff {
		t.Fatal("WithWeakFinalizerEnabled(false) should set finalizerOff=true")
	}

	cfg2 := applyOptions[string, *obj]([]Option[string, *obj]{
		WithWeakFinalizerEnabled[string, obj](true),
	})
	if cfg2.finalizerOff {
		t.Fatal("WithWeakFinalizerEnabled(true) should set finalizerOff=false")
	}
}

func TestApplyListenerClockTickerRunner(t *testing.T) {
	c := &abstractCache[string, int]{}
	c.init(0, 0, nil)
	listener := CacheListenerFunc[string, int](func(string, int) {})
	factory := TickerFactory(func(time.Duration) (<-chan time.Time, Ticker) { return nil, nil })
	runner := func(fn func()) { fn() }

	applyListener(c, listener)
	if c.listener == nil {
		t.Fatal("applyListener failed")
	}

	applyClock(c, time.Now)
	if c.clock == nil {
		t.Fatal("applyClock failed")
	}

	applyTickerFactory(c, factory)
	if c.ticker == nil {
		t.Fatal("applyTickerFactory failed")
	}

	applyRunner(c, runner)
	if c.runner == nil {
		t.Fatal("applyRunner failed")
	}
}

func TestDefaultRunner(t *testing.T) {
	done := make(chan struct{})
	defaultRunner(func() { close(done) })
	<-done
}

func TestNewTicker(t *testing.T) {
	ch, ticker := newTicker(time.Nanosecond)
	ticker.Stop()
	// drain without blocking (ticker may or may not have fired before Stop)
	select {
	case <-ch:
	default:
	}
}
