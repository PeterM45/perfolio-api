package cache

import (
	"sync"
	"time"
)

// Cache defines methods for caching
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, expiration time.Duration)
	Delete(key string)
	Clear()
}

// item represents a cached item
type item struct {
	value      interface{}
	expiration time.Time
}

// InMemoryCache implements an in-memory cache
type InMemoryCache struct {
	items map[string]item
	mu    sync.RWMutex
}

// NewInMemoryCache creates a new in-memory cache
func NewInMemoryCache() *InMemoryCache {
	cache := &InMemoryCache{
		items: make(map[string]item),
	}

	// Run a goroutine to periodically clean up expired items
	go cache.cleanupRoutine()

	return cache
}

// Get retrieves an item from the cache
func (c *InMemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	// Check if the item has expired
	if item.expiration.Before(time.Now()) {
		return nil, false
	}

	return item.value, true
}

// Set adds an item to the cache
func (c *InMemoryCache) Set(key string, value interface{}, expiration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = item{
		value:      value,
		expiration: time.Now().Add(expiration),
	}
}

// Delete removes an item from the cache
func (c *InMemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Clear removes all items from the cache
func (c *InMemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]item)
}

// cleanupRoutine periodically removes expired items
func (c *InMemoryCache) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.items {
			if item.expiration.Before(now) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}
