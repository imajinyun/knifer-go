package cache

import (
	"testing"
	"time"
)

func TestNewFIFOCache(t *testing.T) {
	c := NewFIFOCache[string, int](10)
	if c == nil {
		t.Fatal("NewFIFOCache returned nil")
	}
	if c.Capacity() != 10 {
		t.Fatalf("Capacity = %d, want 10", c.Capacity())
	}
}

func TestNewLFUCache(t *testing.T) {
	c := NewLFUCache[string, int](10)
	if c == nil {
		t.Fatal("NewLFUCache returned nil")
	}
	if c.Capacity() != 10 {
		t.Fatalf("Capacity = %d, want 10", c.Capacity())
	}
}

func TestNewLRUCache(t *testing.T) {
	c := NewLRUCache[string, int](10)
	if c == nil {
		t.Fatal("NewLRUCache returned nil")
	}
	if c.Capacity() != 10 {
		t.Fatalf("Capacity = %d, want 10", c.Capacity())
	}
}

func TestNewLRUCacheWithTimeout(t *testing.T) {
	c := NewLRUCacheWithTimeout[string, int](5, time.Minute)
	if c == nil {
		t.Fatal("NewLRUCacheWithTimeout returned nil")
	}
	if c.Capacity() != 5 {
		t.Fatalf("Capacity = %d, want 5", c.Capacity())
	}
	if c.Timeout() != time.Minute {
		t.Fatalf("Timeout = %v, want 1m", c.Timeout())
	}
}

func TestNewTimedCache(t *testing.T) {
	c := NewTimedCache[string, int](time.Second)
	if c == nil {
		t.Fatal("NewTimedCache returned nil")
	}
	if c.Timeout() != time.Second {
		t.Fatalf("Timeout = %v, want 1s", c.Timeout())
	}
}

func TestNewWeakCache(t *testing.T) {
	c := NewWeakCache[string, struct{ X int }](time.Second)
	if c == nil {
		t.Fatal("NewWeakCache returned nil")
	}
}

func TestNewFIFOWithTimeout(t *testing.T) {
	c := NewFIFOWithTimeout[string, int](5, time.Minute)
	if c == nil {
		t.Fatal("NewFIFOWithTimeout returned nil")
	}
	if c.Capacity() != 5 {
		t.Fatalf("Capacity = %d, want 5", c.Capacity())
	}
	if c.Timeout() != time.Minute {
		t.Fatalf("Timeout = %v, want 1m", c.Timeout())
	}
}

func TestNewLFUWithTimeout(t *testing.T) {
	c := NewLFUWithTimeout[string, int](5, time.Minute)
	if c == nil {
		t.Fatal("NewLFUWithTimeout returned nil")
	}
}

func TestNewTimedScheduled(t *testing.T) {
	c := NewTimedScheduled[string, int](time.Second, 100*time.Millisecond)
	if c == nil {
		t.Fatal("NewTimedScheduled returned nil")
	}
	c.CancelPruneSchedule()
}

func TestAbstractCacheIsFullAndIsEmpty(t *testing.T) {
	c := NewFIFOCache[string, int](2)
	if !c.IsEmpty() {
		t.Fatal("new cache should be empty")
	}
	if c.IsFull() {
		t.Fatal("new cache should not be full")
	}

	c.Put("a", 1)
	if c.IsEmpty() {
		t.Fatal("cache should not be empty after put")
	}
	if c.IsFull() {
		t.Fatal("cache with 1/2 should not be full")
	}

	c.Put("b", 2)
	if !c.IsFull() {
		t.Fatal("cache with 2/2 should be full")
	}
}

func TestCacheObjGetters(t *testing.T) {
	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	obj := newCacheObj("k1", 42, time.Minute, now)

	if got := obj.Key(); got != "k1" {
		t.Fatalf("Key = %q, want %q", got, "k1")
	}
	if got := obj.Value(); got != 42 {
		t.Fatalf("Value = %d, want 42", got)
	}
	if got := obj.TTL(); got != time.Minute {
		t.Fatalf("TTL = %v, want 1m", got)
	}

	last := obj.LastAccess()
	if last.IsZero() {
		t.Fatal("LastAccess should not be zero")
	}

	expTime, ok := obj.ExpiredTime()
	if !ok {
		t.Fatal("ExpiredTime should report ok for non-zero TTL")
	}
	if expTime.Before(now.Add(time.Minute)) {
		t.Fatal("ExpiredTime should be at least now+TTL")
	}
}

func TestCacheObjExpiredTimeNoTTL(t *testing.T) {
	obj := newCacheObj("k", 1, 0, time.Now())
	_, ok := obj.ExpiredTime()
	if ok {
		t.Fatal("ExpiredTime should return false for zero TTL")
	}
}

func TestLFUSetListener(t *testing.T) {
	c := NewLFUCache[string, int](5)
	result := c.SetListener(nil)
	if result == nil {
		t.Fatal("SetListener should return cache for chaining")
	}
}

func TestWeakSetListenerClearPrune(t *testing.T) {
	v := 42
	c := NewWeakCache[string, int](time.Minute)
	c.SetListener(nil)

	c.Put("a", &v)
	if c.Size() != 1 {
		t.Fatalf("size = %d, want 1", c.Size())
	}

	if n := c.Prune(); n != 0 {
		t.Fatalf("Prune on fresh cache = %d, want 0", n)
	}

	c.Clear()
	if c.Size() != 0 {
		t.Fatal("size should be 0 after Clear")
	}
}

func TestWeakCacheHitMissCount(t *testing.T) {
	v := 42
	c := NewWeakCache[string, int](time.Minute)
	c.Put("a", &v)

	if _, ok := c.Get("a"); !ok {
		t.Fatal("Get should find entry")
	}
	if _, ok := c.Get("missing"); ok {
		t.Fatal("Get should miss for missing key")
	}

	if c.HitCount() != 1 {
		t.Fatalf("HitCount = %d, want 1", c.HitCount())
	}
	if c.MissCount() != 1 {
		t.Fatalf("MissCount = %d, want 1", c.MissCount())
	}
}

func TestWeakCacheRemove(t *testing.T) {
	v := 42
	c := NewWeakCache[string, int](time.Minute)
	c.Put("a", &v)
	c.Remove("a")
	if c.Size() != 0 {
		t.Fatal("size should be 0 after Remove")
	}
}
