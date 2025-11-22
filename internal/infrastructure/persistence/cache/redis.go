package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"authentication/internal/infrastructure/observability/metrics"
)

type RedisCache struct {
	client *redis.Client
	name   string
}

func NewRedisCache(client *redis.Client, name string) *RedisCache {
	return &RedisCache{
		client: client,
		name:   name,
	}
}

func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := c.client.Get(ctx, key).Result()
	
	if err == redis.Nil {
		// Cache miss
		metrics.CacheMisses.WithLabelValues(c.name).Inc()
		return ErrCacheMiss
	}
	
	if err != nil {
		metrics.CacheMisses.WithLabelValues(c.name).Inc()
		return err
	}

	// Cache hit
	metrics.CacheHits.WithLabelValues(c.name).Inc()
	
	return json.Unmarshal([]byte(val), dest)
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.client.Exists(ctx, key).Result()
	return result > 0, err
}

var ErrCacheMiss = redis.Nil