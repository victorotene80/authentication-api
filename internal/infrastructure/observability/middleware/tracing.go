package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware returns a Gin middleware that instruments HTTP requests with OpenTelemetry spans.
// For a neobank, this provides critical observability for:
// - Transaction tracing across microservices
// - Performance monitoring of financial operations
// - Audit trail correlation with trace IDs
// - Error tracking and alerting
func TracingMiddleware(tracer trace.Tracer) gin.HandlerFunc {
	// Ensure propagator is configured for distributed tracing
	if otel.GetTextMapPropagator() == nil {
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))
	}

	return func(c *gin.Context) {
		// Skip tracing for health checks and metrics endpoints to reduce noise
		if shouldSkipTracing(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Extract incoming trace context for distributed tracing
		ctx := otel.GetTextMapPropagator().Extract(
			c.Request.Context(),
			propagation.HeaderCarrier(c.Request.Header),
		)

		// Build operation name: HTTP_METHOD /api/path/pattern
		operationName := buildOperationName(c)

		// Start server span with proper attributes
		ctx, span := tracer.Start(ctx, operationName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(buildHTTPAttributes(c)...),
		)
		defer span.End()

		// Inject context into request for downstream propagation
		c.Request = c.Request.WithContext(ctx)

		// Store trace ID in Gin context for structured logging correlation
		traceID := span.SpanContext().TraceID().String()
		spanID := span.SpanContext().SpanID().String()
		c.Set("trace_id", traceID)
		c.Set("span_id", spanID)

		// Add custom headers for client-side trace correlation (useful for debugging)
		c.Header("X-Trace-Id", traceID)

		// Process request
		c.Next()

		// Capture response data
		status := c.Writer.Status()
		span.SetAttributes(
			semconv.HTTPStatusCode(status),
			attribute.Int("http.response.size", c.Writer.Size()),
		)

		// Set span status based on HTTP status code
		setSpanStatus(span, status)

		// Record any errors from Gin context
		if len(c.Errors) > 0 {
			recordErrors(span, c.Errors)
		}

		// Add route pattern if available (important for grouping similar requests)
		if route := c.FullPath(); route != "" {
			span.SetAttributes(semconv.HTTPRoute(route))
		}
	}
}

// buildOperationName creates a meaningful span name
func buildOperationName(c *gin.Context) string {
	method := c.Request.Method

	// Use route pattern if available, otherwise use path
	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}

	return method + " " + path
}

// buildHTTPAttributes creates standard HTTP semantic conventions attributes
func buildHTTPAttributes(c *gin.Context) []attribute.KeyValue {
	req := c.Request

	// Infer scheme
	scheme := req.URL.Scheme
	if scheme == "" {
		if req.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	attrs := []attribute.KeyValue{
		semconv.HTTPMethod(req.Method),
		semconv.HTTPScheme(scheme),
		semconv.HTTPTarget(req.URL.Path),
		semconv.NetHostName(req.Host),
		semconv.HTTPUserAgent(req.UserAgent()),
		semconv.HTTPClientIP(getClientIP(c)),
		attribute.Int64("http.request.size", req.ContentLength),
	}

	// Add query string if present (sanitize sensitive params)
	if query := req.URL.RawQuery; query != "" {
		attrs = append(attrs, attribute.String("http.request.query", sanitizeQuery(query)))
	}

	// Add content type
	if ct := req.Header.Get("Content-Type"); ct != "" {
		attrs = append(attrs, attribute.String("http.request.content_type", ct))
	}

	// Add request ID if present (common in neobank architectures)
	if reqID := c.GetHeader("X-Request-ID"); reqID != "" {
		attrs = append(attrs, attribute.String("http.request.id", reqID))
	}

	return attrs
}

// setSpanStatus sets the span status based on HTTP status code
func setSpanStatus(span trace.Span, statusCode int) {
	switch {
	case statusCode >= 500:
		// Server errors are always span errors
		span.SetStatus(codes.Error, http.StatusText(statusCode))
	case statusCode >= 400:
		// Client errors (4xx) - treat as errors for better alerting
		// In production neobanks, even 4xx can indicate security issues
		span.SetStatus(codes.Error, http.StatusText(statusCode))
	default:
		span.SetStatus(codes.Ok, "")
	}
}

// recordErrors records Gin errors on the span
func recordErrors(span trace.Span, errors []*gin.Error) {
	for _, e := range errors {
		errType := fmt.Sprintf("%v", e.Type)
		span.RecordError(e.Err,
			trace.WithAttributes(
				attribute.String("error.type", errType),
				attribute.String("error.message", e.Error()),
			),
		)
	}
}

// getClientIP extracts the real client IP considering proxies
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header (most common in load-balanced environments)
	if xff := c.Request.Header.Get("X-Forwarded-For"); xff != "" {
		// Take first IP in the chain
		if idx := strings.Index(xff, ","); idx > 0 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := c.Request.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fallback to RemoteAddr
	return c.ClientIP()
}

// sanitizeQuery removes sensitive query parameters from tracing
func sanitizeQuery(query string) string {
	// List of sensitive parameter names common in neobanks
	sensitiveParams := []string{
		"password", "token", "secret", "api_key", "apikey",
		"access_token", "refresh_token", "card_number", "cvv",
		"pin", "ssn", "account_number",
	}

	result := query
	for _, param := range sensitiveParams {
		// Simple sanitization - replace value with [REDACTED]
		for _, prefix := range []string{param + "=", param + "%3D"} {
			if strings.Contains(strings.ToLower(result), strings.ToLower(prefix)) {
				result = strings.ReplaceAll(result, prefix, prefix+"[REDACTED]")
			}
		}
	}
	return result
}

// shouldSkipTracing determines if a path should skip tracing
func shouldSkipTracing(path string) bool {
	skipPaths := []string{
		"/health",
		"/healthz",
		"/ready",
		"/readyz",
		"/livez",
		"/metrics",
		"/ping",
	}

	for _, skip := range skipPaths {
		if path == skip {
			return true
		}
	}
	return false
}
