package tracing

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Tracer is a small abstraction over OpenTelemetry tracer functionality.
// Use this interface across your application so production code depends
// on a small, testable contract instead of the OTEL SDK directly.
type Tracer interface {
	// StartSpan starts a new span with the provided name and options.
	StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span)

	// AddAttributes sets attributes on the span (no-op if span == nil).
	AddAttributes(span trace.Span, attrs ...attribute.KeyValue)

	// AddEvent attaches an event to the span (no-op if span == nil).
	AddEvent(span trace.Span, name string, attrs ...attribute.KeyValue)

	// RecordError records an error on the span (no-op if span == nil).
	RecordError(span trace.Span, err error, attrs ...attribute.KeyValue)

	// Close will flush and shutdown the underlying tracer provider/ exporter.
	Close() error
}

