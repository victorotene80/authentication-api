package tracing

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Tracer interface {
	StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
	AddAttributes(span trace.Span, attrs ...attribute.KeyValue)
	AddEvent(span trace.Span, name string, attrs ...attribute.KeyValue)
	RecordError(span trace.Span, err error, attrs ...attribute.KeyValue)
	Close() error
}
