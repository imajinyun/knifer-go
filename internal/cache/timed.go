package cache

import (
	"sync"
	"time"
)

// TimedCache is an expiration-only cache with no capacity limit.
// Entries are removed only when they expire and are observed by a lookup,
// explicit Prune call, or scheduled prune task.
type TimedCache[K comparable, V any] struct {
	abstractCache[K, V]

	pruneStop chan struct{}
	pruneWG   sync.WaitGroup
}

// NewTimedCache creates a timed cache with timeout as the default TTL.
func NewTimedCache[K comparable, V any](timeout time.Duration) *TimedCache[K, V] {
	return NewTimedCacheWithOptions[K, V](WithTimeout[K, V](timeout))
}

// NewTimedCacheWithOptions creates a timed cache customized by options.
// Capacity is ignored because TimedCache is expiration-only and has no capacity limit.
func NewTimedCacheWithOptions[K comparable, V any](opts ...Option[K, V]) *TimedCache[K, V] {
	return newTimedCacheWithConfig(applyOptions(opts))
}

func newTimedCacheWithConfig[K comparable, V any](cfg cacheConfig[K, V]) *TimedCache[K, V] {
	c := &TimedCache[K, V]{}
	c.init(0, cfg.timeout, timedPrune[K, V])
	applyListener(&c.abstractCache, cfg.listener)
	applyClock(&c.abstractCache, cfg.clock)
	applyTickerFactory(&c.abstractCache, cfg.tickerFactory)
	applyRunner(&c.abstractCache, cfg.runner)
	return c
}

// SetListener sets the removal listener and returns the cache for chaining.
func (c *TimedCache[K, V]) SetListener(l CacheListener[K, V]) Cache[K, V] {
	c.listener = l
	return c
}

func timedPrune[K comparable, V any](c *abstractCache[K, V]) int {
	count := 0
	for _, key := range c.cacheMap.keysInOrder() {
		co, _ := c.cacheMap.get(key)
		if co.isExpired(c.now()) {
			c.removeWithoutLock(key)
			count++
		}
	}
	return count
}

// SchedulePrune starts a background pruning task with delay as the interval.
// The task keeps running until CancelPruneSchedule is called.
func (c *TimedCache[K, V]) SchedulePrune(delay time.Duration) {
	c.pruneStop = make(chan struct{})
	c.pruneWG.Add(1)
	c.run(func() {
		defer c.pruneWG.Done()
		ticks, ticker := c.newTicker(delay)
		defer ticker.Stop()
		for {
			select {
			case <-c.pruneStop:
				return
			case <-ticks:
				c.Prune()
			}
		}
	})
}

// CancelPruneSchedule stops the background pruning task if it is running.
func (c *TimedCache[K, V]) CancelPruneSchedule() {
	if c.pruneStop != nil {
		select {
		case <-c.pruneStop:
			// Already closed.
		default:
			close(c.pruneStop)
		}
		c.pruneWG.Wait()
		c.pruneStop = nil
	}
}
