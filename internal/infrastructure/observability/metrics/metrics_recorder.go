package metrics

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// MetricsRecorder provides convenience methods for recording common metrics
type MetricsRecorder struct {
	// HTTP metrics
	httpRequestCounter    metric.Int64Counter
	httpRequestDuration   metric.Float64Histogram
	httpRequestSizeBytes  metric.Int64Histogram
	httpResponseSizeBytes metric.Int64Histogram

	// Database metrics
	dbQueryDuration metric.Float64Histogram
	dbQueryCounter  metric.Int64Counter

	// Cache metrics
	cacheHitCounter  metric.Int64Counter
	cacheMissCounter metric.Int64Counter

	// Business metrics
	loginAttempts       metric.Int64Counter
	loginSuccesses      metric.Int64Counter
	loginFailures       metric.Int64Counter
	activeUsers         metric.Int64UpDownCounter
	registrationCounter metric.Int64Counter

	// CQRS/Command metrics
	commandExecutionTotal    metric.Int64Counter
	commandExecutionDuration metric.Float64Histogram
	queryExecutionTotal      metric.Int64Counter
	queryExecutionDuration   metric.Float64Histogram

	// Transaction metrics
	transactionTotal    metric.Int64Counter
	transactionDuration metric.Float64Histogram
}

// NewMetricsRecorder creates a new metrics recorder
func NewMetricsRecorder(meter metric.Meter) (*MetricsRecorder, error) {
	mr := &MetricsRecorder{}
	var err error

	// HTTP metrics
	mr.httpRequestCounter, err = meter.Int64Counter(
		"http.server.requests",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request counter: %w", err)
	}

	mr.httpRequestDuration, err = meter.Float64Histogram(
		"http.server.request.duration",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request duration histogram: %w", err)
	}

	mr.httpRequestSizeBytes, err = meter.Int64Histogram(
		"http.server.request.size",
		metric.WithDescription("HTTP request body size in bytes"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request size histogram: %w", err)
	}

	mr.httpResponseSizeBytes, err = meter.Int64Histogram(
		"http.server.response.size",
		metric.WithDescription("HTTP response body size in bytes"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create http response size histogram: %w", err)
	}

	// Database metrics
	mr.dbQueryDuration, err = meter.Float64Histogram(
		"db.query.duration",
		metric.WithDescription("Database query duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create db query duration histogram: %w", err)
	}

	mr.dbQueryCounter, err = meter.Int64Counter(
		"db.queries",
		metric.WithDescription("Total number of database queries"),
		metric.WithUnit("{query}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create db query counter: %w", err)
	}

	// Cache metrics
	mr.cacheHitCounter, err = meter.Int64Counter(
		"cache.hits",
		metric.WithDescription("Total number of cache hits"),
		metric.WithUnit("{hit}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache hit counter: %w", err)
	}

	mr.cacheMissCounter, err = meter.Int64Counter(
		"cache.misses",
		metric.WithDescription("Total number of cache misses"),
		metric.WithUnit("{miss}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache miss counter: %w", err)
	}

	// Business metrics
	mr.loginAttempts, err = meter.Int64Counter(
		"auth.login.attempts",
		metric.WithDescription("Total number of login attempts"),
		metric.WithUnit("{attempt}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create login attempts counter: %w", err)
	}

	mr.loginSuccesses, err = meter.Int64Counter(
		"auth.login.successes",
		metric.WithDescription("Total number of successful logins"),
		metric.WithUnit("{success}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create login successes counter: %w", err)
	}

	mr.loginFailures, err = meter.Int64Counter(
		"auth.login.failures",
		metric.WithDescription("Total number of failed logins"),
		metric.WithUnit("{failure}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create login failures counter: %w", err)
	}

	mr.activeUsers, err = meter.Int64UpDownCounter(
		"auth.active_users",
		metric.WithDescription("Number of currently active users"),
		metric.WithUnit("{user}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create active users counter: %w", err)
	}

	mr.registrationCounter, err = meter.Int64Counter(
		"auth.registrations",
		metric.WithDescription("Total number of user registrations"),
		metric.WithUnit("{registration}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create registration counter: %w", err)
	}

	// CQRS/Command metrics
	mr.commandExecutionTotal, err = meter.Int64Counter(
		"command.execution.total",
		metric.WithDescription("Total number of command executions"),
		metric.WithUnit("{execution}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create command execution counter: %w", err)
	}

	mr.commandExecutionDuration, err = meter.Float64Histogram(
		"command.execution.duration",
		metric.WithDescription("Command execution duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create command execution duration histogram: %w", err)
	}

	mr.queryExecutionTotal, err = meter.Int64Counter(
		"query.execution.total",
		metric.WithDescription("Total number of query executions"),
		metric.WithUnit("{execution}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create query execution counter: %w", err)
	}

	mr.queryExecutionDuration, err = meter.Float64Histogram(
		"query.execution.duration",
		metric.WithDescription("Query execution duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create query execution duration histogram: %w", err)
	}

	// Transaction metrics
	mr.transactionTotal, err = meter.Int64Counter(
		"db.transaction.total",
		metric.WithDescription("Total number of database transactions"),
		metric.WithUnit("{transaction}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction counter: %w", err)
	}

	mr.transactionDuration, err = meter.Float64Histogram(
		"db.transaction.duration",
		metric.WithDescription("Database transaction duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction duration histogram: %w", err)
	}

	return mr, nil
}

// RecordHTTPRequest records metrics for an HTTP request
func (mr *MetricsRecorder) RecordHTTPRequest(ctx context.Context, path, method string, statusCode int, duration time.Duration) {
	attrs := []attribute.KeyValue{
		attribute.String("http.method", method),
		attribute.String("http.route", path),
		attribute.Int("http.status_code", statusCode),
	}

	mr.httpRequestCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	mr.httpRequestDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
}

// RecordHTTPRequestSize records the size of an HTTP request body
func (mr *MetricsRecorder) RecordHTTPRequestSize(ctx context.Context, path, method string, sizeBytes int64) {
	attrs := []attribute.KeyValue{
		attribute.String("http.method", method),
		attribute.String("http.route", path),
	}
	mr.httpRequestSizeBytes.Record(ctx, sizeBytes, metric.WithAttributes(attrs...))
}

// RecordHTTPResponseSize records the size of an HTTP response body
func (mr *MetricsRecorder) RecordHTTPResponseSize(ctx context.Context, path, method string, sizeBytes int64) {
	attrs := []attribute.KeyValue{
		attribute.String("http.method", method),
		attribute.String("http.route", path),
	}
	mr.httpResponseSizeBytes.Record(ctx, sizeBytes, metric.WithAttributes(attrs...))
}

// RecordDatabaseQuery records metrics for a database query
func (mr *MetricsRecorder) RecordDatabaseQuery(ctx context.Context, operation, table string, duration time.Duration) {
	attrs := []attribute.KeyValue{
		attribute.String("db.operation", operation),
		attribute.String("db.table", table),
	}

	mr.dbQueryCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	mr.dbQueryDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
}

// RecordCacheHit records a cache hit
func (mr *MetricsRecorder) RecordCacheHit(ctx context.Context, cacheName string) {
	attrs := []attribute.KeyValue{
		attribute.String("cache.name", cacheName),
	}
	mr.cacheHitCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordCacheMiss records a cache miss
func (mr *MetricsRecorder) RecordCacheMiss(ctx context.Context, cacheName string) {
	attrs := []attribute.KeyValue{
		attribute.String("cache.name", cacheName),
	}
	mr.cacheMissCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordLoginAttempt records a login attempt
func (mr *MetricsRecorder) RecordLoginAttempt(ctx context.Context, username string, success bool) {
	attrs := []attribute.KeyValue{
		attribute.Bool("success", success),
	}

	mr.loginAttempts.Add(ctx, 1, metric.WithAttributes(attrs...))

	if success {
		mr.loginSuccesses.Add(ctx, 1, metric.WithAttributes(attrs...))
	} else {
		mr.loginFailures.Add(ctx, 1, metric.WithAttributes(attrs...))
	}
}

// RecordUserLogin increments active users count
func (mr *MetricsRecorder) RecordUserLogin(ctx context.Context) {
	mr.activeUsers.Add(ctx, 1)
}

// RecordUserLogout decrements active users count
func (mr *MetricsRecorder) RecordUserLogout(ctx context.Context) {
	mr.activeUsers.Add(ctx, -1)
}

// RecordUserRegistration records a new user registration
func (mr *MetricsRecorder) RecordUserRegistration(ctx context.Context, method string) {
	attrs := []attribute.KeyValue{
		attribute.String("registration.method", method),
	}
	mr.registrationCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordCommand records command execution metrics
func (mr *MetricsRecorder) RecordCommand(ctx context.Context, commandName, status string, duration time.Duration) {
	attrs := []attribute.KeyValue{
		attribute.String("command.name", commandName),
		attribute.String("status", status),
	}

	mr.commandExecutionTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
	mr.commandExecutionDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
}

// RecordQuery records query execution metrics
func (mr *MetricsRecorder) RecordQuery(ctx context.Context, queryName, status string, duration time.Duration) {
	attrs := []attribute.KeyValue{
		attribute.String("query.name", queryName),
		attribute.String("status", status),
	}

	mr.queryExecutionTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
	mr.queryExecutionDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
}

// RecordTransaction records transaction metrics
func (mr *MetricsRecorder) RecordTransaction(ctx context.Context, outcome string, duration time.Duration) {
	attrs := []attribute.KeyValue{
		attribute.String("outcome", outcome), // "committed", "rolled_back", "panic"
	}

	mr.transactionTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
	mr.transactionDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
}
