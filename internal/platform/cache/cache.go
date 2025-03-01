package cache

import (
	"time"
)

// Cache defines the interface for cache functionality
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
	Delete(key string)
	Clear()
}

// Factory function to create the right cache implementation
func NewCache(cacheType string, options ...interface{}) Cache {
	switch cacheType {
	case "redis":
		if len(options) > 0 {
			if redisURL, ok := options[0].(string); ok {
				cache, err := NewRedisCache(redisURL)
				if err != nil {
					return NewInMemoryCache(5 * time.Minute) // Fallback to in-memory cache on error
				}
				return cache
			}
		}
		cache, err := NewRedisCache("localhost:6379")
		if err != nil {
			return NewInMemoryCache(5 * time.Minute) // Fallback to in-memory cache on error
		}
		return cache
	}
	return NewInMemoryCache(5 * time.Minute)
}
