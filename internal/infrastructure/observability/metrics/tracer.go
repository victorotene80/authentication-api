package metrics

import (
	"context"
	"crypto/tls"
	"fmt"

	"authentication/shared/config"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// InitTracer initializes OTLP HTTP tracer to send traces to Jaeger
func InitTracer(ctx context.Context, cfg config.TracerConfig, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	// Build OTLP HTTP exporter options
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(cfg.Endpoint),
		otlptracehttp.WithTimeout(cfg.HTTPTimeout),
	}

	// Configure TLS
	if cfg.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	} else {
		opts = append(opts, otlptracehttp.WithTLSClientConfig(&tls.Config{
			MinVersion: tls.VersionTLS12,
		}))
	}

	// Create OTLP HTTP exporter
	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP HTTP exporter: %w", err)
	}

	// Create tracer provider with batching
	tp := sdktrace.NewTracerProvider(
		// Sampling strategy
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SampleRatio)),
		
		// Batch span processor for efficiency
		sdktrace.WithBatcher(exporter,
			sdktrace.WithMaxQueueSize(cfg.MaxQueueSize),
			sdktrace.WithMaxExportBatchSize(cfg.MaxExportBatchSize),
			sdktrace.WithBatchTimeout(cfg.BatchTimeout),
		),
		
		// Resource information
		sdktrace.WithResource(res),
	)

	return tp, nil
}