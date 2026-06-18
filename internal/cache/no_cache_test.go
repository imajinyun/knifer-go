package cache

import (
	"testing"
	"time"
)

func TestNewNoCache(t *testing.T) {
	c := NewNoCache[string, int]()
	if c == nil {
		t.Fatal("NewNoCache returned nil")
	}
}

func TestNoCacheBasic(t *testing.T) {
	c := NewNoCache[string, int]()

	if c.Capacity() != 0 {
		t.Fatalf("Capacity = %d, want 0", c.Capacity())
	}
	if c.Timeout() != 0 {
		t.Fatalf("Timeout = %v, want 0", c.Timeout())
	}
	if c.Size() != 0 {
		t.Fatalf("Size = %d, want 0", c.Size())
	}
	if !c.IsEmpty() {
		t.Fatal("IsEmpty should be true")
	}
	if c.IsFull() {
		t.Fatal("IsFull should be false")
	}
	if c.Prune() != 0 {
		t.Fatal("Prune should return 0")
	}
}

func TestNoCachePutAndGet(t *testing.T) {
	c := NewNoCache[string, int]()
	c.Put("a", 1)
	v, ok := c.Get("a")
	if ok || v != 0 {
		t.Fatalf("Get = (%d, %t), want (0, false)", v, ok)
	}
}

func TestNoCachePutWithTimeout(t *testing.T) {
	c := NewNoCache[string, int]()
	c.PutWithTimeout("a", 1, time.Minute)
	v, ok := c.Get("a")
	if ok || v != 0 {
		t.Fatalf("Get after PutWithTimeout = (%d, %t), want (0, false)", v, ok)
	}
}

func TestNoCacheGetWithUpdate(t *testing.T) {
	c := NewNoCache[string, int]()
	v, ok := c.GetWithUpdate("a", true)
	if ok || v != 0 {
		t.Fatalf("GetWithUpdate = (%d, %t), want (0, false)", v, ok)
	}
}

func TestNoCacheGetOrLoad(t *testing.T) {
	c := NewNoCache[string, int]()
	v, err := c.GetOrLoad("a", func() (int, error) { return 42, nil })
	if err != nil {
		t.Fatalf("GetOrLoad error = %v", err)
	}
	if v != 42 {
		t.Fatalf("GetOrLoad = %d, want 42", v)
	}
}

func TestNoCacheGetOrLoadWith(t *testing.T) {
	c := NewNoCache[string, int]()
	v, err := c.GetOrLoadWith("a", true, time.Minute, func() (int, error) { return 42, nil })
	if err != nil {
		t.Fatalf("GetOrLoadWith error = %v", err)
	}
	if v != 42 {
		t.Fatalf("GetOrLoadWith = %d, want 42", v)
	}
}

func TestNoCacheGetOrLoadNilSupplier(t *testing.T) {
	c := NewNoCache[string, int]()
	v, err := c.GetOrLoad("a", nil)
	if err != nil {
		t.Fatalf("GetOrLoad with nil supplier error = %v", err)
	}
	if v != 0 {
		t.Fatalf("GetOrLoad with nil supplier = %d, want 0", v)
	}
}

func TestNoCacheRemove(t *testing.T) {
	c := NewNoCache[string, int]()
	c.Remove("a")
}

func TestNoCacheContainsKey(t *testing.T) {
	c := NewNoCache[string, int]()
	if c.ContainsKey("a") {
		t.Fatal("ContainsKey should be false")
	}
}

func TestNoCacheClear(t *testing.T) {
	c := NewNoCache[string, int]()
	c.Clear()
}

func TestNoCacheKeysAndValues(t *testing.T) {
	c := NewNoCache[string, int]()
	if k := c.Keys(); k != nil {
		t.Fatalf("Keys = %v, want nil", k)
	}
	if v := c.Values(); v != nil {
		t.Fatalf("Values = %v, want nil", v)
	}
}

func TestNoCacheSetListener(t *testing.T) {
	c := NewNoCache[string, int]()
	c2 := c.SetListener(nil)
	if c2 == nil {
		t.Fatal("SetListener should not return nil")
	}
	// Verify returned value is actually a NoCache
	if _, ok := c2.(NoCache[string, int]); !ok {
		t.Fatal("SetListener should return NoCache")
	}
}

func TestNoCacheHitMissCount(t *testing.T) {
	c := NewNoCache[string, int]()
	if c.HitCount() != 0 {
		t.Fatalf("HitCount = %d, want 0", c.HitCount())
	}
	if c.MissCount() != 0 {
		t.Fatalf("MissCount = %d, want 0", c.MissCount())
	}
}