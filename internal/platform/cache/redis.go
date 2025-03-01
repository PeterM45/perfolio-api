package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisCache is a Redis-based cache implementation
type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(redisURL string) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})

	// Check if the client is able to connect to Redis
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client: client,
		ctx:    context.Background(),
	}, nil
}

// Get retrieves a value from Redis
func (c *RedisCache) Get(key string) (interface{}, bool) {
	val, err := c.client.Get(c.ctx, key).Result()
	if err == redis.Nil {
		return nil, false
	} else if err != nil {
		return nil, false
	}

	var result interface{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, false
	}

	return result, true
}

// Set adds a value to Redis
func (c *RedisCache) Set(key string, value interface{}, ttl time.Duration) {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return
	}

	c.client.Set(c.ctx, key, jsonValue, ttl)
}

// Delete removes a value from Redis
func (c *RedisCache) Delete(key string) {
	c.client.Del(c.ctx, key)
}

// Clear removes all values with a specific prefix
// Note: Redis doesn't have a direct "clear all" for a namespace without using KEYS
// which is not recommended for production
func (c *RedisCache) Clear() {
	// For full clear, this would be
	// c.client.FlushAll(c.ctx)
	// But that's too destructive for shared Redis instances

	// In a real application, consider using key prefixes and scanning
	// for more selective clearing
}
