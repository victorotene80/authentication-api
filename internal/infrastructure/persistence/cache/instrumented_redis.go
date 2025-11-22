// internal/infrastructure/persistence/cache/instrumented_cache.go
package cache

import (
	"authentication/internal/application/contracts"
	"authentication/internal/infrastructure/observability/metrics"
	"authentication/shared/utils"
	"context"
	"time"
)

// InstrumentedCache wraps a Cache with metrics collection
type InstrumentedCache struct {
	wrapped   contracts.Cache
	cacheName string
}

func NewInstrumentedCache(wrapped contracts.Cache, cacheName string) contracts.Cache {
	return &InstrumentedCache{
		wrapped:   wrapped,
		cacheName: cacheName,
	}
}

func (i *InstrumentedCache) Get(ctx context.Context, key string, dest interface{}) error {
	start := utils.NowUTC()
	err := i.wrapped.Get(ctx, key, dest)
	duration := time.Since(start).Seconds()

	status := "success"
	if err != nil {
		if err == ErrCacheMiss {
			metrics.CacheMisses.WithLabelValues(i.cacheName).Inc()
			status = "miss"
		} else {
			status = "error"
		}
	} else {
		metrics.CacheHits.WithLabelValues(i.cacheName).Inc()
	}

	metrics.CacheOperationDuration.WithLabelValues("get", i.cacheName).Observe(duration)
	metrics.CacheOperationTotal.WithLabelValues("get", i.cacheName, status).Inc()

	return err
}

func (i *InstrumentedCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	start := utils.NowUTC()
	err := i.wrapped.Set(ctx, key, value, ttl)
	duration := time.Since(start).Seconds()

	status := "success"
	if err != nil {
		status = "error"
	}

	metrics.CacheOperationDuration.WithLabelValues("set", i.cacheName).Observe(duration)
	metrics.CacheOperationTotal.WithLabelValues("set", i.cacheName, status).Inc()

	return err
}

func (i *InstrumentedCache) Delete(ctx context.Context, key string) error {
	start := utils.NowUTC()
	err := i.wrapped.Delete(ctx, key)
	duration := time.Since(start).Seconds()

	status := "success"
	if err != nil {
		status = "error"
	}

	metrics.CacheOperationDuration.WithLabelValues("delete", i.cacheName).Observe(duration)
	metrics.CacheOperationTotal.WithLabelValues("delete", i.cacheName, status).Inc()

	return err
}

func (i *InstrumentedCache) Exists(ctx context.Context, key string) (bool, error) {
	start := utils.NowUTC()
	exists, err := i.wrapped.Exists(ctx, key)
	duration := time.Since(start).Seconds()

	status := "success"
	if err != nil {
		status = "error"
	}

	metrics.CacheOperationDuration.WithLabelValues("exists", i.cacheName).Observe(duration)
	metrics.CacheOperationTotal.WithLabelValues("exists", i.cacheName, status).Inc()

	return exists, err
}
