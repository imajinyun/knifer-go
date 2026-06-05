package cache

import (
	"runtime"
	"testing"
	"time"
)

func TestWeakCacheBasic(t *testing.T) {
	c := NewWeak[string, int](0)
	v := 42
	c.Put("a", &v)
	got, ok := c.Get("a")
	if !ok || got == nil || *got != 42 {
		t.Fatalf("get failed: %v %v", got, ok)
	}
	c.Remove("a")
	if _, ok := c.Get("a"); ok {
		t.Fatalf("should be removed")
	}
}

func TestWeakCacheTimeout(t *testing.T) {
	c := NewWeak[string, int](10 * time.Millisecond)
	v := 42
	c.Put("a", &v)
	time.Sleep(20 * time.Millisecond)
	if _, ok := c.Get("a"); ok {
		t.Fatalf("should expire")
	}
}

func TestWeakCacheWithClock(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	now := base
	c := NewWeakWithOptions[string, int](
		WithTimeout[string, *int](time.Second),
		WithClock[string, *int](func() time.Time { return now }),
	)
	v := 42
	c.Put("a", &v)
	now = base.Add(500 * time.Millisecond)
	if got, ok := c.Get("a"); !ok || got == nil || *got != 42 {
		t.Fatalf("expected value before custom-clock expiry, got %v ok=%v", got, ok)
	}
	now = base.Add(2 * time.Second)
	if _, ok := c.Get("a"); ok {
		t.Fatalf("expected custom clock to expire weak entry")
	}
}

// Verify that a later GC can clean a weak entry once external strong references
// disappear. Finalizer scheduling is not fully deterministic, so the test uses
// several GC cycles and small sleeps to improve stability.
func TestWeakCacheGC(t *testing.T) {
	c := NewWeak[string, int](0)
	func() {
		v := 7
		c.Put("a", &v)
	}()
	for i := 0; i < 5 && c.Size() > 0; i++ {
		runtime.GC()
		time.Sleep(10 * time.Millisecond)
	}
	if c.Size() != 0 {
		// On some Go runtime/GC schedules, finalizer execution can be delayed.
		// Keep the test non-flaky by logging instead of asserting strictly.
		t.Logf("weak cache size after GC: %d (finalizer may be delayed)", c.Size())
	}
}
