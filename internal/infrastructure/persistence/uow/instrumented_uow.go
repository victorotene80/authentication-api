package uow

import (
	"context"
	"database/sql"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"authentication/shared/logging"
	"authentication/internal/infrastructure/observability/metrics"
	uow "authentication/shared/persistence"
	"authentication/shared/tracing"
)

// InstrumentedUnitOfWork wraps a UnitOfWork with observability
type InstrumentedUnitOfWork struct {
	wrapped uow.UnitOfWork
	logger  logging.Logger
	tracer  tracing.Tracer
	dbName  string
	env     string
}

// NewInstrumentedUnitOfWork creates a new instrumented unit of work with observability
func NewInstrumentedUnitOfWork(
	base uow.UnitOfWork,
	logger logging.Logger,
	tracer tracing.Tracer,
	dbName string,
	env string,
) uow.UnitOfWork {
	return &InstrumentedUnitOfWork{
		wrapped: base,
		logger:  logger.With(zap.String("component", "unit_of_work")),
		tracer:  tracer,
		dbName:  dbName,
		env:     env,
	}
}

// Begin starts a new transaction with observability
func (i *InstrumentedUnitOfWork) Begin(ctx context.Context) error {
	ctx, span := i.tracer.StartSpan(ctx, "uow.begin", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	i.tracer.AddAttributes(span,
		attribute.String("db.operation", "begin"),
		attribute.String("db.name", i.dbName),
	)

	i.logger.Debug(ctx, "Beginning transaction",
		zap.String("db_name", i.dbName),
	)

	start := time.Now()
	err := i.wrapped.Begin(ctx)
	duration := time.Since(start)

	if err != nil {
		i.logger.Error(ctx, "Failed to begin transaction",
			zap.Error(err),
			zap.Duration("duration", duration),
		)
		i.tracer.RecordError(span, err,
			attribute.String("error.type", "begin_failed"),
		)
		span.SetStatus(codes.Error, "begin failed")
		return err
	}

	i.logger.Debug(ctx, "Transaction started",
		zap.Duration("duration", duration),
	)

	i.tracer.AddAttributes(span,
		attribute.Float64("duration_ms", float64(duration.Milliseconds())),
	)
	i.tracer.AddEvent(span, "transaction.started")
	span.SetStatus(codes.Ok, "started")

	return nil
}

// Commit commits the current transaction with observability
func (i *InstrumentedUnitOfWork) Commit(ctx context.Context) error {
	ctx, span := i.tracer.StartSpan(ctx, "uow.commit", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	i.tracer.AddAttributes(span,
		attribute.String("db.operation", "commit"),
		attribute.String("db.name", i.dbName),
	)

	i.logger.Debug(ctx, "Committing transaction",
		zap.String("db_name", i.dbName),
	)

	start := time.Now()
	err := i.wrapped.Commit(ctx)
	duration := time.Since(start)

	if err != nil {
		i.logger.Error(ctx, "Failed to commit transaction",
			zap.Error(err),
			zap.Duration("duration", duration),
		)
		i.tracer.RecordError(span, err,
			attribute.String("error.type", "commit_failed"),
		)
		span.SetStatus(codes.Error, "commit failed")
		return err
	}

	i.logger.Debug(ctx, "Transaction committed",
		zap.Duration("duration", duration),
	)

	i.tracer.AddAttributes(span,
		attribute.Float64("duration_ms", float64(duration.Milliseconds())),
	)
	i.tracer.AddEvent(span, "transaction.committed")
	span.SetStatus(codes.Ok, "committed")

	return nil
}

// Rollback rolls back the current transaction with observability
func (i *InstrumentedUnitOfWork) Rollback(ctx context.Context) error {
	ctx, span := i.tracer.StartSpan(ctx, "uow.rollback", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	i.tracer.AddAttributes(span,
		attribute.String("db.operation", "rollback"),
		attribute.String("db.name", i.dbName),
	)

	i.logger.Warn(ctx, "Rolling back transaction",
		zap.String("db_name", i.dbName),
	)

	start := time.Now()
	err := i.wrapped.Rollback(ctx)
	duration := time.Since(start)

	if err != nil {
		i.logger.Error(ctx, "Failed to rollback transaction",
			zap.Error(err),
			zap.Duration("duration", duration),
		)
		i.tracer.RecordError(span, err,
			attribute.String("error.type", "rollback_failed"),
		)
		span.SetStatus(codes.Error, "rollback failed")
		return err
	}

	i.logger.Debug(ctx, "Transaction rolled back",
		zap.Duration("duration", duration),
	)

	i.tracer.AddAttributes(span,
		attribute.Float64("duration_ms", float64(duration.Milliseconds())),
	)
	i.tracer.AddEvent(span, "transaction.rolled_back")
	span.SetStatus(codes.Ok, "rolled back")

	return nil
}

// IsInTransaction returns true if a transaction is active
func (i *InstrumentedUnitOfWork) IsInTransaction() bool {
	return i.wrapped.IsInTransaction()
}

// GetTx returns the current transaction
func (i *InstrumentedUnitOfWork) GetTx() *sql.Tx {
	return i.wrapped.GetTx()
}

// Execute runs a function within a transaction with full observability
func (i *InstrumentedUnitOfWork) Execute(
	ctx context.Context,
	fn func(ctx context.Context, tx *sql.Tx) error,
) error {
	ctx, span := i.tracer.StartSpan(ctx, "uow.execute", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	i.tracer.AddAttributes(span,
		attribute.String("db.operation", "transaction"),
		attribute.String("db.name", i.dbName),
		attribute.String("db.system", "postgresql"),
	)

	i.logger.Debug(ctx, "Executing transaction",
		zap.String("db_name", i.dbName),
	)

	start := time.Now()
	err := i.wrapped.Execute(ctx, fn)
	duration := time.Since(start).Seconds()

	// Determine transaction status
	status := "committed"
	if err != nil {
		status = "rolled_back"
		
		i.logger.Error(ctx, "Transaction failed and rolled back",
			zap.Error(err),
			zap.Duration("duration", time.Duration(duration*float64(time.Second))),
		)

		i.tracer.RecordError(span, err,
			attribute.String("transaction.status", "rolled_back"),
		)
		span.SetStatus(codes.Error, "transaction failed")
	} else {
		i.logger.Debug(ctx, "Transaction completed successfully",
			zap.Duration("duration", time.Duration(duration*float64(time.Second))),
		)

		i.tracer.AddAttributes(span,
			attribute.String("transaction.status", "committed"),
		)
		span.SetStatus(codes.Ok, "committed")
	}

	// Record metrics
	metrics.DatabaseTransactionDuration.Observe(duration)
	metrics.DatabaseTransactionsTotal.WithLabelValues(status).Inc()

	// Add transaction summary to trace
	i.tracer.AddAttributes(span,
		attribute.Float64("duration_ms", duration*1000),
		attribute.String("status", status),
	)

	// Add event for transaction completion
	i.tracer.AddEvent(span, "transaction.completed",
		attribute.String("status", status),
		attribute.Float64("duration_ms", duration*1000),
	)

	// Warn if transaction took too long (e.g., > 1 second)
	if duration > 1.0 {
		i.logger.Warn(ctx, "Slow transaction detected",
			zap.Float64("duration_seconds", duration),
			zap.String("status", status),
		)

		i.tracer.AddEvent(span, "transaction.slow",
			attribute.Float64("duration_seconds", duration),
		)
	}

	return err
}