package cache

import (
	"sync"
	"time"
)

// item represents a cached item with expiration
type item struct {
	value      interface{}
	expiration int64
}

// InMemoryCache is a simple in-memory cache implementation
type InMemoryCache struct {
	items map[string]item
	mu    sync.RWMutex
	ttl   time.Duration
}

// NewInMemoryCache creates a new in-memory cache
func NewInMemoryCache(defaultTTL time.Duration) *InMemoryCache {
	cache := &InMemoryCache{
		items: make(map[string]item),
		ttl:   defaultTTL,
	}
	go cache.cleanupRoutine()
	return cache
}

// Get retrieves a value from the cache
func (c *InMemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	// Check if item has expired
	if item.expiration > 0 && item.expiration < time.Now().UnixNano() {
		return nil, false
	}

	return item.value, true
}

// Set adds a value to the cache
func (c *InMemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	var expiration int64

	if ttl == 0 {
		ttl = c.ttl
	}

	if ttl > 0 {
		expiration = time.Now().Add(ttl).UnixNano()
	}

	c.mu.Lock()
	c.items[key] = item{
		value:      value,
		expiration: expiration,
	}
	c.mu.Unlock()
}

// Delete removes a value from the cache
func (c *InMemoryCache) Delete(key string) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

// Clear removes all values from the cache
func (c *InMemoryCache) Clear() {
	c.mu.Lock()
	c.items = make(map[string]item)
	c.mu.Unlock()
}

// cleanupRoutine periodically cleans up expired items
func (c *InMemoryCache) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now().UnixNano()

		c.mu.Lock()
		for k, v := range c.items {
			if v.expiration > 0 && v.expiration < now {
				delete(c.items, k)
			}
		}
		c.mu.Unlock()
	}
}
