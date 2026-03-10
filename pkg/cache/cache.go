package cache

import (
	"sync"
	"time"
)

type entry[V any] struct {
	value     V
	expiresAt time.Time
}

type Cache[K comparable, V any] struct {
	mu      sync.RWMutex
	items   map[K]entry[V]
	maxSize int
	stop    chan struct{}
	stopped bool
}

func New[K comparable, V any](cleanupInterval time.Duration, maxSize int) *Cache[K, V] {
	c := &Cache[K, V]{
		items:   make(map[K]entry[V]),
		maxSize: maxSize,
		stop:    make(chan struct{}),
	}

	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.evictExpired()
			case <-c.stop:
				return
			}
		}
	}()

	return c
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.items[key]
	if !ok || time.Now().After(e.expiresAt) {
		var zero V
		return zero, false
	}
	return e.value, true
}

func (c *Cache[K, V]) Set(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.items[key]; !exists && c.maxSize > 0 && len(c.items) >= c.maxSize {
		c.evictOldest()
	}

	c.items[key] = entry[V]{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
}

func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

func (c *Cache[K, V]) evictExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for k, e := range c.items {
		if now.After(e.expiresAt) {
			delete(c.items, k)
		}
	}
}

// evictOldest removes the entry closest to expiration. Must be called with mu held.
func (c *Cache[K, V]) evictOldest() {
	var oldestKey K
	var oldestTime time.Time
	first := true

	for k, e := range c.items {
		if first || e.expiresAt.Before(oldestTime) {
			oldestKey = k
			oldestTime = e.expiresAt
			first = false
		}
	}

	if !first {
		delete(c.items, oldestKey)
	}
}

func (c *Cache[K, V]) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.stopped {
		close(c.stop)
		c.stopped = true
	}
}
