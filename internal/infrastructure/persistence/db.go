package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"authentication/shared/config"
	"authentication/shared/logging"
	"authentication/shared/tracing"
)

// NewDatabase creates a new database connection with observability
func NewDatabase(ctx context.Context, cfg config.DatabaseConfig, logger logging.Logger, tracer tracing.Tracer) (*sql.DB, error) {
	ctx, span := tracer.StartSpan(ctx, "database.connect", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	// Add connection attributes to trace
	tracer.AddAttributes(span,
		attribute.String("db.system", cfg.Driver),
		attribute.String("db.name", cfg.Name),
		attribute.String("db.host", cfg.Host),
		attribute.Int("db.port", cfg.Port),
	)

	logger.Info(ctx, "Initializing database connection",
		zap.String("driver", cfg.Driver),
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Name),
		zap.String("ssl_mode", cfg.SSLMode),
	)

	// Build DSN
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	// Open database connection
	db, err := sql.Open(cfg.Driver, dsn)
	if err != nil {
		logger.Error(ctx, "Failed to open database connection",
			zap.Error(err),
			zap.String("driver", cfg.Driver),
		)
		tracer.RecordError(span, err,
			attribute.String("error.type", "connection_failed"),
		)
		span.SetStatus(codes.Error, "failed to open database")
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Apply connection pool settings
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	logger.Debug(ctx, "Database connection pool configured",
		zap.Int("max_open_conns", cfg.MaxOpenConns),
		zap.Int("max_idle_conns", cfg.MaxIdleConns),
		zap.Duration("conn_max_lifetime", cfg.ConnMaxLifetime),
		zap.Duration("conn_max_idle_time", cfg.ConnMaxIdleTime),
	)

	tracer.AddAttributes(span,
		attribute.Int("db.pool.max_open", cfg.MaxOpenConns),
		attribute.Int("db.pool.max_idle", cfg.MaxIdleConns),
	)

	// Verify connection with timeout
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pingStart := time.Now()
	if err := db.PingContext(pingCtx); err != nil {
		logger.Error(ctx, "Failed to ping database",
			zap.Error(err),
			zap.Duration("timeout", 5*time.Second),
		)
		tracer.RecordError(span, err,
			attribute.String("error.type", "ping_failed"),
		)
		span.SetStatus(codes.Error, "failed to ping database")

		// Attempt to close the connection
		if closeErr := db.Close(); closeErr != nil {
			logger.Warn(ctx, "Failed to close database after ping failure",
				zap.Error(closeErr),
			)
		}

		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	pingDuration := time.Since(pingStart)

	logger.Info(ctx, "Database connection established successfully",
		zap.Duration("ping_duration", pingDuration),
		zap.Int("max_open_conns", cfg.MaxOpenConns),
	)

	// Add success event to trace
	tracer.AddEvent(span, "database.connected",
		attribute.String("db.name", cfg.Name),
		attribute.Float64("ping_duration_ms", float64(pingDuration.Milliseconds())),
	)
	span.SetStatus(codes.Ok, "connected")

	return db, nil
}

// CloseDatabase gracefully closes the database connection
func CloseDatabase(ctx context.Context, db *sql.DB, logger logging.Logger, tracer tracing.Tracer) {
	if db == nil {
		return
	}

	ctx, span := tracer.StartSpan(ctx, "database.close")
	defer span.End()

	logger.Info(ctx, "Closing database connection")

	// Get final stats before closing
	stats := db.Stats()
	logger.Debug(ctx, "Final database connection pool stats",
		zap.Int("open_connections", stats.OpenConnections),
		zap.Int("in_use", stats.InUse),
		zap.Int("idle", stats.Idle),
		zap.Int64("wait_count", stats.WaitCount),
		zap.Duration("wait_duration", stats.WaitDuration),
	)

	tracer.AddAttributes(span,
		attribute.Int("db.pool.open_connections", stats.OpenConnections),
		attribute.Int("db.pool.in_use", stats.InUse),
		attribute.Int("db.pool.idle", stats.Idle),
	)

	if err := db.Close(); err != nil {
		logger.Error(ctx, "Failed to close database connection",
			zap.Error(err),
		)
		tracer.RecordError(span, err,
			attribute.String("error.type", "close_failed"),
		)
		span.SetStatus(codes.Error, "failed to close")
		return
	}

	logger.Info(ctx, "Database connection closed successfully")
	tracer.AddEvent(span, "database.closed")
	span.SetStatus(codes.Ok, "closed")
}

// PingDatabase checks database connectivity
func PingDatabase(ctx context.Context, db *sql.DB, logger logging.Logger, tracer tracing.Tracer) error {
	ctx, span := tracer.StartSpan(ctx, "database.ping")
	defer span.End()

	start := time.Now()

	if err := db.PingContext(ctx); err != nil {
		logger.Error(ctx, "Database ping failed",
			zap.Error(err),
			zap.Duration("duration", time.Since(start)),
		)
		tracer.RecordError(span, err,
			attribute.String("error.type", "ping_failed"),
		)
		span.SetStatus(codes.Error, "ping failed")
		return err
	}

	duration := time.Since(start)
	logger.Debug(ctx, "Database ping successful",
		zap.Duration("duration", duration),
	)

	tracer.AddAttributes(span,
		attribute.Float64("ping_duration_ms", float64(duration.Milliseconds())),
	)
	span.SetStatus(codes.Ok, "success")

	return nil
}

// GetDatabaseStats returns current database connection pool statistics
func GetDatabaseStats(ctx context.Context, db *sql.DB, logger logging.Logger, tracer tracing.Tracer) sql.DBStats {
	ctx, span := tracer.StartSpan(ctx, "database.get_stats")
	defer span.End()

	stats := db.Stats()

	logger.Debug(ctx, "Retrieved database connection pool stats",
		zap.Int("open_connections", stats.OpenConnections),
		zap.Int("in_use", stats.InUse),
		zap.Int("idle", stats.Idle),
		zap.Int64("wait_count", stats.WaitCount),
		zap.Duration("wait_duration", stats.WaitDuration),
		zap.Int64("max_idle_closed", stats.MaxIdleClosed),
		zap.Int64("max_lifetime_closed", stats.MaxLifetimeClosed),
	)

	tracer.AddAttributes(span,
		attribute.Int("db.pool.open", stats.OpenConnections),
		attribute.Int("db.pool.in_use", stats.InUse),
		attribute.Int("db.pool.idle", stats.Idle),
		attribute.Int64("db.pool.wait_count", stats.WaitCount),
	)

	span.SetStatus(codes.Ok, "retrieved")

	return stats
}

// WaitForDatabase waits for database to become available with retries
func WaitForDatabase(ctx context.Context, db *sql.DB, logger logging.Logger, tracer tracing.Tracer, maxRetries int, retryDelay time.Duration) error {
	ctx, span := tracer.StartSpan(ctx, "database.wait_ready")
	defer span.End()

	tracer.AddAttributes(span,
		attribute.Int("max_retries", maxRetries),
		attribute.Float64("retry_delay_seconds", retryDelay.Seconds()),
	)

	logger.Info(ctx, "Waiting for database to become available",
		zap.Int("max_retries", maxRetries),
		zap.Duration("retry_delay", retryDelay),
	)

	for i := 0; i < maxRetries; i++ {
		tracer.AddEvent(span, "retry_attempt",
			attribute.Int("attempt", i+1),
		)

		if err := db.PingContext(ctx); err == nil {
			logger.Info(ctx, "Database is available",
				zap.Int("attempts", i+1),
			)
			tracer.AddAttributes(span,
				attribute.Int("successful_attempt", i+1),
			)
			span.SetStatus(codes.Ok, "database ready")
			return nil
		}

		if i < maxRetries-1 {
			logger.Warn(ctx, "Database not available, retrying",
				zap.Int("attempt", i+1),
				zap.Int("max_retries", maxRetries),
				zap.Duration("retry_in", retryDelay),
			)

			select {
			case <-ctx.Done():
				logger.Error(ctx, "Context cancelled while waiting for database",
					zap.Error(ctx.Err()),
				)
				tracer.RecordError(span, ctx.Err(),
					attribute.String("error.type", "context_cancelled"),
				)
				span.SetStatus(codes.Error, "context cancelled")
				return ctx.Err()
			case <-time.After(retryDelay):
				// Delay finished, continue to next retry attempt
				logger.Debug(ctx, "Retry delay elapsed, retrying database connection",
					zap.Duration("retry_delay", retryDelay),
				)

				tracer.AddEvent(span, "retry_delay_elapsed",
					attribute.Float64("retry_delay_ms", float64(retryDelay.Milliseconds())),
				)
			}
		}
	}

	err := fmt.Errorf("database not available after %d attempts", maxRetries)
	logger.Error(ctx, "Failed to connect to database",
		zap.Error(err),
		zap.Int("attempts", maxRetries),
	)
	tracer.RecordError(span, err,
		attribute.String("error.type", "max_retries_exceeded"),
		attribute.Int("max_retries", maxRetries),
	)
	span.SetStatus(codes.Error, "max retries exceeded")

	return err
}
