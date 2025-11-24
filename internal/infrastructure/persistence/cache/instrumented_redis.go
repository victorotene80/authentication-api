package cache

import (
	"context"
	"time"

	"authentication/internal/application/contracts"
	"authentication/internal/infrastructure/observability/metrics"
	"authentication/shared/tracing"
	"authentication/shared/utils"

	"go.opentelemetry.io/otel/attribute"
)

type InstrumentedCache struct {
	wrapped   contracts.Cache
	cacheName string
	metrics   *metrics.MetricsRecorder
	tracer    tracing.Tracer
}

func NewInstrumentedCache(
	wrapped contracts.Cache,
	cacheName string,
	metricsRecorder *metrics.MetricsRecorder,
	tracer tracing.Tracer,
) contracts.Cache {
	return &InstrumentedCache{
		wrapped:   wrapped,
		cacheName: cacheName,
		metrics:   metricsRecorder,
		tracer:    tracer,
	}
}

func (i *InstrumentedCache) Get(ctx context.Context, key string, dest interface{}) error {
	ctx, span := i.tracer.StartSpan(ctx, "cache.get")
	defer span.End()

	i.tracer.AddAttributes(span,
		attribute.String("cache.name", i.cacheName),
		attribute.String("cache.key", key),
	)

	start := utils.NowUTC()
	err := i.wrapped.Get(ctx, key, dest)
	duration := time.Since(start)

	status := "success"
	if err != nil {
		if err == ErrCacheMiss {
			status = "miss"
			if i.metrics != nil {
				i.metrics.RecordCacheMiss(ctx, i.cacheName)
			}
			i.tracer.AddEvent(span, "cache.miss")
		} else {
			status = "error"
			i.tracer.RecordError(span, err)
		}
	} else {
		if i.metrics != nil {
			i.metrics.RecordCacheHit(ctx, i.cacheName)
		}
		i.tracer.AddEvent(span, "cache.hit")
	}

	if i.metrics != nil {
		// Record duration as a database-style metric (optional)
		i.metrics.RecordDatabaseQuery(ctx, "cache.get."+i.cacheName, "duration", duration)
	}

	i.tracer.AddAttributes(span,
		attribute.String("status", status),
		attribute.Float64("duration_ms", duration.Seconds()*1000),
	)

	return err
}

func (i *InstrumentedCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	ctx, span := i.tracer.StartSpan(ctx, "cache.set")
	defer span.End()

	i.tracer.AddAttributes(span,
		attribute.String("cache.name", i.cacheName),
		attribute.String("cache.key", key),
	)

	start := utils.NowUTC()
	err := i.wrapped.Set(ctx, key, value, ttl)
	duration := time.Since(start)

	status := "success"
	if err != nil {
		status = "error"
		i.tracer.RecordError(span, err)
	}

	i.tracer.AddAttributes(span,
		attribute.String("status", status),
		attribute.Float64("duration_ms", duration.Seconds()*1000),
	)

	return err
}

func (i *InstrumentedCache) Delete(ctx context.Context, key string) error {
	ctx, span := i.tracer.StartSpan(ctx, "cache.delete")
	defer span.End()

	i.tracer.AddAttributes(span,
		attribute.String("cache.name", i.cacheName),
		attribute.String("cache.key", key),
	)

	start := utils.NowUTC()
	err := i.wrapped.Delete(ctx, key)
	duration := time.Since(start)

	status := "success"
	if err != nil {
		status = "error"
		i.tracer.RecordError(span, err)
	}

	i.tracer.AddAttributes(span,
		attribute.String("status", status),
		attribute.Float64("duration_ms", duration.Seconds()*1000),
	)

	return err
}

func (i *InstrumentedCache) Exists(ctx context.Context, key string) (bool, error) {
	ctx, span := i.tracer.StartSpan(ctx, "cache.exists")
	defer span.End()

	i.tracer.AddAttributes(span,
		attribute.String("cache.name", i.cacheName),
		attribute.String("cache.key", key),
	)

	start := utils.NowUTC()
	exists, err := i.wrapped.Exists(ctx, key)
	duration := time.Since(start)

	status := "success"
	if err != nil {
		status = "error"
		i.tracer.RecordError(span, err)
	}

	i.tracer.AddAttributes(span,
		attribute.String("status", status),
		attribute.Float64("duration_ms", duration.Seconds()*1000),
	)

	return exists, err
}
