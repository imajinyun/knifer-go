package cache

import (
	"testing"
	"time"
)

func TestNoCache(t *testing.T) {
	c := NewNo[string, int]()
	c.Put("a", 1)
	if c.Size() != 0 || !c.IsEmpty() {
		t.Fatalf("nocache size/empty wrong")
	}
	if _, ok := c.Get("a"); ok {
		t.Fatalf("nocache get hit?")
	}
	v, err := c.GetOrLoad("k", func() (int, error) { return 7, nil })
	if err != nil || v != 7 {
		t.Fatalf("nocache loader: %v %v", v, err)
	}
}

func TestNoCacheExtraMethods(t *testing.T) {
	c := NewNo[string, int]()
	if c.HitCount() != 0 || c.MissCount() != 0 {
		t.Fatalf("nocache stats: hit=%d miss=%d", c.HitCount(), c.MissCount())
	}
	c.SetListener(nil)
	c.Clear()
	if c.Keys() != nil || c.Values() != nil || c.ContainsKey("x") {
		t.Fatalf("nocache unexpected returns")
	}
}

func TestNoCacheCapacityAndTimeout(t *testing.T) {
	c := NewNo[string, int]()
	if c.Capacity() != 0 {
		t.Fatalf("Capacity = %d, want 0", c.Capacity())
	}
	if c.Timeout() != 0 {
		t.Fatalf("Timeout = %v, want 0", c.Timeout())
	}
}

func TestNoCachePutWithTimeout(t *testing.T) {
	c := NewNo[string, int]()
	c.PutWithTimeout("a", 1, time.Second) // should be no-op
	if c.Size() != 0 {
		t.Fatal("PutWithTimeout should not change size")
	}
}

func TestNoCacheGetWithUpdate(t *testing.T) {
	c := NewNo[string, int]()
	c.Put("a", 1)
	v, ok := c.GetWithUpdate("a", true)
	if ok || v != 0 {
		t.Fatalf("GetWithUpdate = (%d, %v), want (0, false)", v, ok)
	}
}

func TestNoCacheGetOrLoadNilSupplier(t *testing.T) {
	c := NewNo[string, int]()
	v, err := c.GetOrLoad("x", nil)
	if err != nil || v != 0 {
		t.Fatalf("GetOrLoad nil supplier = (%d, %v), want (0, nil)", v, err)
	}
}

func TestNoCacheGetOrLoadWith(t *testing.T) {
	c := NewNo[string, int]()
	v, err := c.GetOrLoadWith("x", false, 0, func() (int, error) { return 42, nil })
	if err != nil || v != 42 {
		t.Fatalf("GetOrLoadWith = (%d, %v), want (42, nil)", v, err)
	}
}

func TestNoCacheRemoveIsFullPrune(t *testing.T) {
	c := NewNo[string, int]()
	c.Remove("a") // no-op
	if c.IsFull() {
		t.Fatal("IsFull should be false")
	}
	if c.Prune() != 0 {
		t.Fatal("Prune should return 0")
	}
	if !c.IsEmpty() {
		t.Fatal("IsEmpty should be true (always empty for NoCache)")
	}
}
