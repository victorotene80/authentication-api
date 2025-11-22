package tracing

import (
	"authentication/shared/config"
	"context"
	"fmt"
	"io"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

var globalTracer trace.Tracer

// InitTracer initializes the global tracer using OTLP exporter
func InitTracer(cfg config.TracerConfig) (io.Closer, error) {
	if !cfg.Enabled {
		return &noopCloser{}, nil
	}

	// Create OTLP HTTP exporter
	exporter, err := otlptrace.New(
		context.Background(),
		otlptracehttp.NewClient(
			otlptracehttp.WithEndpoint(cfg.JaegerEndpoint), // or collector endpoint
			otlptracehttp.WithInsecure(),                   // disable TLS for local dev
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource with service information
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			attribute.String("environment", cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)

	// Set context propagation
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	globalTracer = tp.Tracer(cfg.ServiceName)

	return &tracerCloser{tp: tp}, nil
}


// GetTracer returns the global tracer
func GetTracer() trace.Tracer {
	if globalTracer == nil {
		// Return a no-op tracer if not initialized
		return otel.Tracer("noop")
	}
	return globalTracer
}

// StartSpan starts a new span with the given name
func StartSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return GetTracer().Start(ctx, spanName, opts...)
}

// AddSpanAttributes adds attributes to the current span
func AddSpanAttributes(span trace.Span, attrs ...attribute.KeyValue) {
	if span != nil {
		span.SetAttributes(attrs...)
	}
}

// AddSpanEvent adds an event to the current span
func AddSpanEvent(span trace.Span, name string, attrs ...attribute.KeyValue) {
	if span != nil {
		span.AddEvent(name, trace.WithAttributes(attrs...))
	}
}

// RecordError records an error in the current span
func RecordError(span trace.Span, err error, attrs ...attribute.KeyValue) {
	if span != nil && err != nil {
		span.RecordError(err, trace.WithAttributes(attrs...))
	}
}

// tracerCloser implements io.Closer for graceful shutdown
type tracerCloser struct {
	tp *sdktrace.TracerProvider
}

func (tc *tracerCloser) Close() error {
	if tc.tp != nil {
		return tc.tp.Shutdown(context.Background())
	}
	return nil
}

// noopCloser is a no-op closer
type noopCloser struct{}

func (nc *noopCloser) Close() error {
	return nil
}

// Helper functions for common trace attributes

// UserAttributes returns common user-related attributes
func UserAttributes(userID, email string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("user.id", userID),
		attribute.String("user.email", email),
	}
}

// DatabaseAttributes returns common database-related attributes
func DatabaseAttributes(operation, table string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("db.operation", operation),
		attribute.String("db.table", table),
	}
}

// HTTPAttributes returns common HTTP-related attributes
func HTTPAttributes(method, path string, statusCode int) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("http.method", method),
		attribute.String("http.path", path),
		attribute.Int("http.status_code", statusCode),
	}
}

// CommandAttributes returns command-related attributes
func CommandAttributes(commandType string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("command.type", commandType),
	}
}

// QueryAttributes returns query-related attributes
func QueryAttributes(queryType string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("query.type", queryType),
	}
}

// EventAttributes returns event-related attributes
func EventAttributes(eventType string, aggregateID string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("event.type", eventType),
		attribute.String("event.aggregate_id", aggregateID),
	}
}
