package cache

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Core scenarios from hutool-cache CacheTest.

func TestFIFOCache(t *testing.T) {
	var removedKey, removedValue string
	c := NewFIFO[string, string](3)
	c.SetListener(CacheListenerFunc[string, string](func(k, v string) {
		removedKey, removedValue = k, v
	}))
	c.PutWithTimeout("key1", "value1", 3*time.Second)
	c.PutWithTimeout("key2", "value2", 3*time.Second)
	c.PutWithTimeout("key3", "value3", 3*time.Second)
	c.PutWithTimeout("key4", "value4", 3*time.Second)

	// Adding the 4th entry should evict the oldest key, key1.
	if v, ok := c.Get("key1"); ok {
		t.Fatalf("key1 should be evicted, got %q", v)
	}
	if removedKey != "key1" || removedValue != "value1" {
		t.Fatalf("listener got: %s=%s", removedKey, removedValue)
	}
}

func TestFIFOCapacity(t *testing.T) {
	c := NewFIFO[string, string](100)
	for i := 0; i < 500; i++ {
		c.Put(itoa(i), "v")
	}
	if got := c.Size(); got != 100 {
		t.Fatalf("size: %d", got)
	}
}

func TestLFUCache(t *testing.T) {
	c := NewLFU[string, string](3)
	c.PutWithTimeout("key1", "value1", 3*time.Second)
	c.Get("key1") // Increase the access count by 1.
	c.PutWithTimeout("key2", "value2", 3*time.Second)
	c.PutWithTimeout("key3", "value3", 3*time.Second)
	c.PutWithTimeout("key4", "value4", 3*time.Second)

	if _, ok := c.Get("key1"); !ok {
		t.Fatalf("key1 should still exist")
	}
	if _, ok := c.Get("key2"); ok {
		t.Fatalf("key2 should be evicted")
	}
	if _, ok := c.Get("key3"); ok {
		t.Fatalf("key3 should be evicted")
	}
}

func TestLRUCache(t *testing.T) {
	c := NewLRU[string, string](3)
	c.PutWithTimeout("key1", "value1", 3*time.Second)
	c.PutWithTimeout("key2", "value2", 3*time.Second)
	c.PutWithTimeout("key3", "value3", 3*time.Second)
	c.Get("key1") // Move key1 to the tail; key2 becomes least recently used.
	c.PutWithTimeout("key4", "value4", 3*time.Second)

	if _, ok := c.Get("key1"); !ok {
		t.Fatalf("key1 should still exist")
	}
	if _, ok := c.Get("key2"); ok {
		t.Fatalf("key2 should be evicted (LRU)")
	}
}

func TestLRURemoveCount(t *testing.T) {
	var count int32
	c := NewLRUWithTimeout[string, int](3, 1*time.Millisecond)
	c.SetListener(CacheListenerFunc[string, int](func(string, int) {
		atomic.AddInt32(&count, 1)
	}))
	for i := 0; i < 10; i++ {
		c.Put("key-"+itoa(i), i)
		// Sleep between puts so the previous value expires and prune triggers onRemove.
		time.Sleep(2 * time.Millisecond)
	}
	if c.Size() != 1 {
		// With ttl=1ms and a 2ms sleep, each put-triggered prune removes old
		// expired entries, leaving only the last inserted entry.
		t.Fatalf("expected size=1, got %d", c.Size())
	}
}

func TestTimedCache(t *testing.T) {
	c := NewTimed[string, string](4 * time.Millisecond)
	c.PutWithTimeout("key1", "value1", 1*time.Millisecond)
	c.PutWithTimeout("key2", "value2", 5*time.Second)
	c.Put("key3", "value3")               // Uses the default 4ms timeout.
	c.PutWithTimeout("key4", "value4", 0) // Never expires.

	c.SchedulePrune(5 * time.Millisecond)
	defer c.CancelPruneSchedule()
	time.Sleep(20 * time.Millisecond)

	if _, ok := c.Get("key1"); ok {
		t.Fatalf("key1 should expire")
	}
	if v, ok := c.Get("key2"); !ok || v != "value2" {
		t.Fatalf("key2: %v %v", v, ok)
	}
	if _, ok := c.Get("key3"); ok {
		t.Fatalf("key3 should expire")
	}
	if v, ok := c.Get("key4"); !ok || v != "value4" {
		t.Fatalf("key4: %v %v", v, ok)
	}

	v, err := c.GetOrLoad("key3", func() (string, error) { return "Default supplier", nil })
	if err != nil || v != "Default supplier" {
		t.Fatalf("GetOrLoad: %v %v", v, err)
	}
}

// Mirrors hutool whenContainsKeyTimeout_shouldCallOnRemove.
func TestContainsKeyExpiredOnRemove(t *testing.T) {
	timeout := 50 * time.Millisecond
	c := NewTimed[int, string](timeout)
	var counter int32
	c.SetListener(CacheListenerFunc[int, string](func(int, string) {
		atomic.AddInt32(&counter, 1)
	}))
	c.Put(1, "value1")
	time.Sleep(100 * time.Millisecond)
	if c.ContainsKey(1) {
		t.Fatalf("should not contain key 1")
	}
	if got := atomic.LoadInt32(&counter); got != 1 {
		t.Fatalf("listener counter: %d", got)
	}
}

// Mirrors hutool reentrantCache_clear_Method_Test.
func TestLRUClearTriggersListener(t *testing.T) {
	var removeCount int32
	c := NewLRU[string, string](4)
	c.SetListener(CacheListenerFunc[string, string](func(string, string) {
		atomic.AddInt32(&removeCount, 1)
	}))
	c.Put("key1", "String1")
	c.Put("key2", "String2")
	c.Put("key3", "String3")
	c.Put("key1", "String4") // Replacement triggers one removal notification.
	c.Put("key4", "String5")
	c.Clear() // Clearing triggers the remaining 4 removal notifications.
	if got := atomic.LoadInt32(&removeCount); got != 5 {
		t.Fatalf("removeCount expected 5, got %d", got)
	}
}

func TestGetOrLoad(t *testing.T) {
	c := NewLRU[string, int](3)
	v, err := c.GetOrLoad("a", func() (int, error) { return 42, nil })
	if err != nil || v != 42 {
		t.Fatalf("first: %v %v", v, err)
	}
	// The second call hits the cache directly and does not call supplier again.
	called := 0
	v, err = c.GetOrLoad("a", func() (int, error) {
		called++
		return 99, nil
	})
	if err != nil || v != 42 || called != 0 {
		t.Fatalf("second: v=%d err=%v called=%d", v, err, called)
	}
}

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

func TestLRUReadWriteConcurrency(t *testing.T) {
	const N = 10
	c := NewLRU[int, int](N)
	for i := 0; i < N; i++ {
		c.Put(i, i)
	}
	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				c.Get(idx)
			}
		}(i)
	}
	wg.Wait()
	// The order should still be 0..9. Each get moves a node to the tail, so 0 is
	// the first element after iterating over all keys in ascending order.
	got := ""
	for i := 0; i < N; i++ {
		if v, ok := c.Get(i); ok {
			got += itoa(v)
		} else {
			got += "x"
		}
	}
	if got != "0123456789" {
		t.Fatalf("got: %s", got)
	}
	// Adding 11 should evict 0, which is now the least recently used entry.
	c.Put(11, 11)
	if _, ok := c.Get(0); ok {
		t.Fatalf("key 0 should be evicted")
	}
}

func TestRemoveAndContains(t *testing.T) {
	c := NewLRU[string, int](5)
	c.Put("a", 1)
	c.Put("b", 2)
	if !c.ContainsKey("a") {
		t.Fatal("expected contains a")
	}
	c.Remove("a")
	if c.ContainsKey("a") {
		t.Fatal("a should be removed")
	}
	if c.Size() != 1 {
		t.Fatalf("size: %d", c.Size())
	}
}

func TestHitMissCount(t *testing.T) {
	c := NewLRU[string, int](5)
	c.Put("a", 1)
	c.Get("a")
	c.Get("a")
	c.Get("b")
	if c.HitCount() != 2 || c.MissCount() != 1 {
		t.Fatalf("hit=%d miss=%d", c.HitCount(), c.MissCount())
	}
}

func itoa(i int) string {
	// Simple int-to-string conversion used by tests to avoid extra dependencies.
	if i == 0 {
		return "0"
	}
	neg := false
	if i < 0 {
		neg = true
		i = -i
	}
	buf := [20]byte{}
	n := 0
	for i > 0 {
		buf[n] = byte('0' + i%10)
		i /= 10
		n++
	}
	if neg {
		buf[n] = '-'
		n++
	}
	// Reverse the digits in place.
	for j, k := 0, n-1; j < k; j, k = j+1, k-1 {
		buf[j], buf[k] = buf[k], buf[j]
	}
	return string(buf[:n])
}
