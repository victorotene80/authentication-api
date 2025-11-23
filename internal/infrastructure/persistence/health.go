package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"

	"authentication/shared/logging"
	"authentication/shared/tracing"
)

// HealthChecker performs database health checks
type HealthChecker struct {
	db     *sql.DB
	logger logging.Logger
	tracer tracing.Tracer
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(db *sql.DB, logger logging.Logger, tracer tracing.Tracer) *HealthChecker {
	return &HealthChecker{
		db:     db,
		logger: logger,
		tracer: tracer,
	}
}

// HealthStatus represents the health status of the database
type HealthStatus struct {
	Status            string        `json:"status"`
	ResponseTime      time.Duration `json:"response_time_ms"`
	OpenConnections   int           `json:"open_connections"`
	InUse             int           `json:"in_use"`
	Idle              int           `json:"idle"`
	WaitCount         int64         `json:"wait_count"`
	WaitDuration      time.Duration `json:"wait_duration_ms"`
	MaxIdleClosed     int64         `json:"max_idle_closed"`
	MaxLifetimeClosed int64         `json:"max_lifetime_closed"`
}

// Check performs a basic health check (liveness probe)
func (h *HealthChecker) Check(ctx context.Context) (*HealthStatus, error) {
	ctx, span := h.tracer.StartSpan(ctx, "database.health_check")
	defer span.End()

	h.tracer.AddAttributes(span,
		attribute.String("check.type", "liveness"),
	)

	start := time.Now()

	h.logger.Debug(ctx, "Starting database health check")

	// Ping database
	if err := h.db.PingContext(ctx); err != nil {
		responseTime := time.Since(start)
		
		h.logger.Error(ctx, "Database health check failed",
			zap.Error(err),
			zap.Duration("response_time", responseTime),
		)
		
		h.tracer.RecordError(span, err,
			attribute.String("check.type", "liveness"),
			attribute.String("error.type", "ping_failed"),
		)
		span.SetStatus(codes.Error, "health check failed")

		return &HealthStatus{
			Status:       "unhealthy",
			ResponseTime: responseTime,
		}, fmt.Errorf("database ping failed: %w", err)
	}

	responseTime := time.Since(start)

	// Get connection pool stats
	stats := h.db.Stats()

	healthStatus := &HealthStatus{
		Status:            "healthy",
		ResponseTime:      responseTime,
		OpenConnections:   stats.OpenConnections,
		InUse:             stats.InUse,
		Idle:              stats.Idle,
		WaitCount:         stats.WaitCount,
		WaitDuration:      stats.WaitDuration,
		MaxIdleClosed:     stats.MaxIdleClosed,
		MaxLifetimeClosed: stats.MaxLifetimeClosed,
	}

	h.logger.Debug(ctx, "Database health check passed",
		zap.Duration("response_time", responseTime),
		zap.Int("open_connections", stats.OpenConnections),
		zap.Int("in_use", stats.InUse),
		zap.Int("idle", stats.Idle),
	)

	h.tracer.AddAttributes(span,
		attribute.Float64("response_time_ms", float64(responseTime.Milliseconds())),
		attribute.Int("db.pool.open", stats.OpenConnections),
		attribute.Int("db.pool.in_use", stats.InUse),
		attribute.Int("db.pool.idle", stats.Idle),
	)

	// Warn if connection pool is stressed
	if stats.OpenConnections > 0 {
		utilization := float64(stats.InUse) / float64(stats.OpenConnections)
		
		if utilization > 0.8 {
			h.logger.Warn(ctx, "Database connection pool is under pressure",
				zap.Int("in_use", stats.InUse),
				zap.Int("open_connections", stats.OpenConnections),
				zap.Float64("utilization", utilization),
			)
			
			h.tracer.AddEvent(span, "pool.high_utilization",
				attribute.Float64("utilization", utilization),
			)
		}
	}

	h.tracer.AddEvent(span, "health_check.completed",
		attribute.String("status", "healthy"),
	)
	span.SetStatus(codes.Ok, "healthy")

	return healthStatus, nil
}

// CheckReadiness performs a thorough check for readiness probe
func (h *HealthChecker) CheckReadiness(ctx context.Context) error {
	ctx, span := h.tracer.StartSpan(ctx, "database.readiness_check")
	defer span.End()

	h.tracer.AddAttributes(span,
		attribute.String("check.type", "readiness"),
	)

	h.logger.Debug(ctx, "Starting database readiness check")

	// Create timeout context
	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	start := time.Now()

	// Try to execute a simple query
	var result int
	err := h.db.QueryRowContext(checkCtx, "SELECT 1").Scan(&result)
	
	duration := time.Since(start)

	if err != nil {
		h.logger.Error(ctx, "Database readiness check failed",
			zap.Error(err),
			zap.Duration("duration", duration),
		)
		h.tracer.RecordError(span, err,
			attribute.String("check.type", "readiness"),
			attribute.String("error.type", "query_failed"),
		)
		span.SetStatus(codes.Error, "readiness check failed")
		return fmt.Errorf("readiness check failed: %w", err)
	}

	if result != 1 {
		err := fmt.Errorf("unexpected query result: %d", result)
		h.logger.Error(ctx, "Database readiness check returned unexpected result",
			zap.Error(err),
			zap.Int("result", result),
		)
		h.tracer.RecordError(span, err,
			attribute.String("error.type", "unexpected_result"),
		)
		span.SetStatus(codes.Error, "unexpected result")
		return err
	}

	h.logger.Debug(ctx, "Database readiness check passed",
		zap.Duration("duration", duration),
	)

	h.tracer.AddAttributes(span,
		attribute.Float64("duration_ms", float64(duration.Milliseconds())),
	)
	h.tracer.AddEvent(span, "readiness_check.completed",
		attribute.String("status", "ready"),
	)
	span.SetStatus(codes.Ok, "ready")

	return nil
}

// CheckDeep performs a comprehensive database check
func (h *HealthChecker) CheckDeep(ctx context.Context) error {
	ctx, span := h.tracer.StartSpan(ctx, "database.deep_check")
	defer span.End()

	h.tracer.AddAttributes(span,
		attribute.String("check.type", "deep"),
	)

	h.logger.Info(ctx, "Starting deep database health check")

	// 1. Check basic connectivity
	h.tracer.AddEvent(span, "check.connectivity")
	if err := h.CheckReadiness(ctx); err != nil {
		h.tracer.RecordError(span, err,
			attribute.String("phase", "connectivity"),
		)
		span.SetStatus(codes.Error, "connectivity check failed")
		return fmt.Errorf("basic connectivity check failed: %w", err)
	}

	// 2. Check connection pool health
	h.tracer.AddEvent(span, "check.pool_health")
	stats := h.db.Stats()
	
	h.tracer.AddAttributes(span,
		attribute.Int("db.pool.open", stats.OpenConnections),
		attribute.Int("db.pool.in_use", stats.InUse),
		attribute.Int("db.pool.idle", stats.Idle),
		attribute.Int64("db.pool.wait_count", stats.WaitCount),
	)
	
	// Warn if too many connections waiting
	if stats.WaitCount > 1000 {
		h.logger.Warn(ctx, "High connection wait count detected",
			zap.Int64("wait_count", stats.WaitCount),
			zap.Duration("total_wait_duration", stats.WaitDuration),
		)
		h.tracer.AddEvent(span, "pool.high_wait_count",
			attribute.Int64("wait_count", stats.WaitCount),
		)
	}

	// Warn if many connections closed due to idle
	if stats.MaxIdleClosed > 100 {
		h.logger.Warn(ctx, "Many idle connections closed",
			zap.Int64("max_idle_closed", stats.MaxIdleClosed),
		)
		h.tracer.AddEvent(span, "pool.high_idle_closed",
			attribute.Int64("max_idle_closed", stats.MaxIdleClosed),
		)
	}

	// 3. Test transaction capability
	h.tracer.AddEvent(span, "check.transaction")
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		h.logger.Error(ctx, "Failed to begin transaction in deep check",
			zap.Error(err),
		)
		h.tracer.RecordError(span, err,
			attribute.String("phase", "transaction_begin"),
		)
		span.SetStatus(codes.Error, "transaction test failed")
		return fmt.Errorf("transaction test failed: %w", err)
	}

	if err := tx.Rollback(); err != nil {
		h.logger.Error(ctx, "Failed to rollback transaction in deep check",
			zap.Error(err),
		)
		h.tracer.RecordError(span, err,
			attribute.String("phase", "transaction_rollback"),
		)
		span.SetStatus(codes.Error, "transaction rollback failed")
		return fmt.Errorf("transaction rollback failed: %w", err)
	}

	h.logger.Info(ctx, "Deep database health check completed successfully",
		zap.Int("open_connections", stats.OpenConnections),
		zap.Int("in_use", stats.InUse),
		zap.Int("idle", stats.Idle),
	)

	h.tracer.AddEvent(span, "deep_check.completed")
	span.SetStatus(codes.Ok, "all checks passed")

	return nil
}

// MonitorHealth continuously monitors database health
func (h *HealthChecker) MonitorHealth(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	h.logger.Info(ctx, "Starting database health monitor",
		zap.Duration("interval", interval),
	)

	for {
		select {
		case <-ctx.Done():
			h.logger.Info(ctx, "Database health monitor shutting down",
				zap.Error(ctx.Err()),
			)
			return

		case <-ticker.C:
			checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			
			status, err := h.Check(checkCtx)
			if err != nil {
				h.logger.Error(checkCtx, "Health monitor check failed",
					zap.Error(err),
				)
			} else if status.ResponseTime > 100*time.Millisecond {
				h.logger.Warn(checkCtx, "Database health check response time is slow",
					zap.Duration("response_time", status.ResponseTime),
				)
			}
			
			cancel()
		}
	}
}

// GetConnectionPoolUtilization returns the current connection pool utilization percentage
func (h *HealthChecker) GetConnectionPoolUtilization(ctx context.Context) float64 {
	ctx, span := h.tracer.StartSpan(ctx, "database.get_pool_utilization")
	defer span.End()

	stats := h.db.Stats()
	
	if stats.OpenConnections == 0 {
		span.SetStatus(codes.Ok, "no connections")
		return 0
	}

	utilization := float64(stats.InUse) / float64(stats.OpenConnections)
	
	h.logger.Debug(ctx, "Connection pool utilization",
		zap.Float64("utilization", utilization),
		zap.Int("in_use", stats.InUse),
		zap.Int("open_connections", stats.OpenConnections),
	)

	h.tracer.AddAttributes(span,
		attribute.Float64("utilization", utilization),
		attribute.Int("in_use", stats.InUse),
		attribute.Int("open", stats.OpenConnections),
	)
	span.SetStatus(codes.Ok, "calculated")

	return utilization
}