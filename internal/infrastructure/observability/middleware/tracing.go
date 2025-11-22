package middleware

import (
	"authentication/internal/infrastructure/observability/tracing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware creates OpenTelemetry spans for incoming HTTP requests.
// It safely handles unmatched routes, propagates context, and records standard attributes.
func TracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Determine route path (handle unmatched routes safely)
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Start a new span for this request
		ctx, span := tracing.StartSpan(
			c.Request.Context(),
			c.Request.Method+" "+path,
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		// Inject trace ID into context for logging correlation
		traceID := span.SpanContext().TraceID().String()
		c.Set("trace_id", traceID)

		// Update request context
		c.Request = c.Request.WithContext(ctx)

		// Record standard HTTP request attributes (using OpenTelemetry conventions)
		span.SetAttributes(
			semconv.HTTPMethodKey.String(c.Request.Method),
			semconv.HTTPTargetKey.String(c.Request.URL.Path),
			semconv.HTTPURLKey.String(c.Request.URL.String()),
			//semconv.HTTPHostKey.String(c.Request.Host),
			semconv.HTTPUserAgentKey.String(c.Request.UserAgent()),
			semconv.HTTPClientIPKey.String(c.ClientIP()),
			semconv.HTTPSchemeKey.String(c.Request.URL.Scheme),
		)

		// Continue to the next middleware/handler
		c.Next()

		// Add response-related attributes
		statusCode := c.Writer.Status()
		span.SetAttributes(
			semconv.HTTPStatusCodeKey.Int(statusCode),
		)

		// Set span status based on response code
		if statusCode >= 400 {
			span.SetStatus(codes.Error, "HTTP request failed")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		

		// Record any Gin errors attached to the context
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				tracing.RecordError(span, err.Err)
			}
		}

		// Add final event marker for trace observability
		tracing.AddSpanEvent(span, "Request completed")
	}
}
