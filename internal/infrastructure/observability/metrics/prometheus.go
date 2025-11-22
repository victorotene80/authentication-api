package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP Metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "endpoint"},
	)

	HTTPRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 7),
		},
		[]string{"method", "endpoint"},
	)

	HTTPResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 7),
		},
		[]string{"method", "endpoint"},
	)

	// Business Metrics
	UserRegistrations = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_user_registrations_total",
			Help: "Total number of user registrations",
		},
	)

	UserRegistrationErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_user_registration_errors_total",
			Help: "Total number of user registration errors",
		},
		[]string{"error_type"},
	)

	UserLoginAttempts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_user_login_attempts_total",
			Help: "Total number of login attempts",
		},
		[]string{"status"}, // success, failed, blocked
	)

	UserLoginDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "auth_user_login_duration_seconds",
			Help:    "User login operation duration",
			Buckets: []float64{.01, .025, .05, .1, .25, .5, 1, 2.5, 5},
		},
	)

	ActiveSessions = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "auth_active_sessions_total",
			Help: "Current number of active sessions",
		},
	)

	SessionCreations = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_session_creations_total",
			Help: "Total number of sessions created",
		},
	)

	SessionTerminations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_session_terminations_total",
			Help: "Total number of sessions terminated",
		},
		[]string{"reason"}, // logout, timeout, revoked
	)

	// Database Metrics
	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_database_query_duration_seconds",
			Help:    "Database query latency in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5},
		},
		[]string{"operation", "table", "db_name", "env"},
	)

	DatabaseQueryTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_database_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table", "status", "db_name", "env"},
	)

	DatabaseConnectionPoolSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "auth_database_connection_pool_size",
			Help: "Current database connection pool size",
		},
		[]string{"type", "state", "db_name", "env"},
	)

	DatabaseTransactionDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "auth_database_transaction_duration_seconds",
			Help:    "Database transaction duration",
			Buckets: []float64{.01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
	)

	DatabaseTransactionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_database_transactions_total",
			Help: "Total number of database transactions",
		},
		[]string{"status"}, // committed, rolled_back
	)

	// Command/Query Bus Metrics
	CommandExecutionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_command_execution_duration_seconds",
			Help:    "Command execution duration in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
		},
		[]string{"command_type"},
	)

	CommandExecutionTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_command_executions_total",
			Help: "Total number of command executions",
		},
		[]string{"command_type", "status"},
	)

	QueryExecutionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_query_execution_duration_seconds",
			Help:    "Query execution duration in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"query_type"},
	)

	QueryExecutionTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_query_executions_total",
			Help: "Total number of query executions",
		},
		[]string{"query_type", "status"},
	)

	// Event/Message Bus Metrics
	EventsPublished = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_events_published_total",
			Help: "Total number of events published",
		},
		[]string{"event_type", "status"},
	)

	EventsProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_events_processed_total",
			Help: "Total number of events processed",
		},
		[]string{"event_type", "status"},
	)

	EventProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_event_processing_duration_seconds",
			Help:    "Event processing duration",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5},
		},
		[]string{"event_type"},
	)

	MessageQueueSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "auth_message_queue_size",
			Help: "Current message queue size",
		},
		[]string{"queue_name"},
	)

	MessagePublishDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_message_publish_duration_seconds",
			Help:    "Message publish duration",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5},
		},
		[]string{"queue_name"},
	)

	// Cache Metrics
	CacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_name"},
	)

	CacheMisses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_name"},
	)

	CacheOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_cache_operation_duration_seconds",
			Help:    "Cache operation duration",
			Buckets: []float64{.0001, .0005, .001, .005, .01, .025, .05, .1},
		},
		[]string{"operation", "cache_name"},
	)

	CacheSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "auth_cache_size_bytes",
			Help: "Current cache size in bytes",
		},
		[]string{"cache_name"},
	)

	CacheEvictions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_cache_evictions_total",
			Help: "Total number of cache evictions",
		},
		[]string{"cache_name"},
	)

	CacheOperationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_operation_total",
			Help: "Total cache operations by result (success/error/miss)",
		},
		[]string{"operation", "cache", "status"},
	)



	// Outbox Pattern Metrics
	OutboxMessagesProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_outbox_messages_processed_total",
			Help: "Total number of outbox messages processed",
		},
		[]string{"status"},
	)

	OutboxQueueSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "auth_outbox_queue_size",
			Help: "Current number of pending outbox messages",
		},
	)

	OutboxProcessingDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "auth_outbox_processing_duration_seconds",
			Help:    "Outbox message processing duration",
			Buckets: []float64{.01, .025, .05, .1, .25, .5, 1, 2.5, 5},
		},
	)

	OutboxRetries = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_outbox_retries_total",
			Help: "Total number of outbox message retries",
		},
	)

	// JWT Token Metrics
	JWTTokensIssued = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_jwt_tokens_issued_total",
			Help: "Total number of JWT tokens issued",
		},
	)

	JWTTokenValidations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_jwt_token_validations_total",
			Help: "Total number of JWT token validations",
		},
		[]string{"status"}, // valid, invalid, expired
	)

	JWTTokenRefreshes = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_jwt_token_refreshes_total",
			Help: "Total number of JWT token refreshes",
		},
	)

	JWTTokenRevocations = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_jwt_token_revocations_total",
			Help: "Total number of JWT token revocations",
		},
	)

	// Validation Metrics
	ValidationErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_validation_errors_total",
			Help: "Total number of validation errors",
		},
		[]string{"field", "error_type"},
	)

	// Rate Limiting Metrics
	RateLimitHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_rate_limit_hits_total",
			Help: "Total number of rate limit hits",
		},
		[]string{"endpoint"}, //"user_id"}
	)

	// Password Metrics
	PasswordHashingDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "auth_password_hashing_duration_seconds",
			Help:    "Password hashing operation duration",
			Buckets: []float64{.01, .025, .05, .1, .25, .5, 1},
		},
	)

	PasswordResets = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_password_resets_total",
			Help: "Total number of password reset requests",
		},
	)

	// System Metrics
	SystemErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_system_errors_total",
			Help: "Total number of system errors",
		},
		[]string{"component", "error_type"},
	)

	GoroutineCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "auth_goroutines_count",
			Help: "Current number of goroutines",
		},
	)

	MemoryUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "auth_memory_usage_bytes",
			Help: "Memory usage in bytes",
		},
		[]string{"type"}, 
	)
)