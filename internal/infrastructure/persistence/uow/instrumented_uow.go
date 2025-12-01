package uow

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	uow "authentication/internal/application/contracts/persistence"
	"authentication/internal/infrastructure/observability/metrics"
	"authentication/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type instrumentedUnitOfWork struct {
	*unitOfWork
	logger  *zap.Logger
	tracer  tracing.Tracer
	metrics *metrics.MetricsRecorder
}

func NewInstrumentedUnitOfWork(
	db *sql.DB,
	logger *zap.Logger,
	tracer tracing.Tracer,
	metricsRecorder *metrics.MetricsRecorder,
) uow.UnitOfWork {
	return &instrumentedUnitOfWork{
		unitOfWork: &unitOfWork{db: db},
		logger:     logger.With(zap.String("component", "unit_of_work")),
		tracer:     tracer,
		metrics:    metricsRecorder,
	}
}

func (u *instrumentedUnitOfWork) Begin(ctx context.Context) error {
	start := time.Now()
	ctx, span := u.tracer.StartSpan(ctx, "uow.Begin",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("db.operation", "begin_transaction")),
	)
	defer span.End()

	u.logger.Debug("beginning transaction")
	err := u.unitOfWork.Begin(ctx)
	duration := time.Since(start)

	if err != nil {
		u.logger.Error("failed to begin transaction", zap.Error(err), zap.Duration("duration", duration))
		span.RecordError(err)
		u.tracer.AddAttributes(span,
			attribute.String("error.type", fmt.Sprintf("%T", err)),
			attribute.String("error.message", err.Error()),
		)
		//span.SetStatus(codes.Error, "failed to begin transaction")
	} else {
		u.logger.Debug("transaction started successfully", zap.Duration("duration", duration))
		span.SetStatus(codes.Ok, "success")
	}

	span.SetAttributes(attribute.Float64("duration_ms", duration.Seconds()*1000))
	if u.metrics != nil {
		u.metrics.RecordDatabaseQuery(ctx, "BEGIN", "transaction", duration)
	}

	return err
}

func (u *instrumentedUnitOfWork) Commit(ctx context.Context) error {
	start := time.Now()
	ctx, span := u.tracer.StartSpan(ctx, "uow.Commit",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("db.operation", "commit_transaction")),
	)
	defer span.End()

	u.logger.Debug("committing transaction")
	err := u.unitOfWork.Commit(ctx)
	duration := time.Since(start)

	if err != nil {
		u.logger.Error("failed to commit transaction", zap.Error(err), zap.Duration("duration", duration))
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to commit")
	} else {
		u.logger.Debug("transaction committed successfully", zap.Duration("duration", duration))
		span.SetStatus(codes.Ok, "success")
	}

	span.SetAttributes(attribute.Float64("duration_ms", duration.Seconds()*1000))
	if u.metrics != nil {
		u.metrics.RecordDatabaseQuery(ctx, "COMMIT", "transaction", duration)
	}

	return err
}

func (u *instrumentedUnitOfWork) Rollback(ctx context.Context) error {
	start := time.Now()
	ctx, span := u.tracer.StartSpan(ctx, "uow.Rollback",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("db.operation", "rollback_transaction")),
	)
	defer span.End()

	u.logger.Debug("rolling back transaction")
	err := u.unitOfWork.Rollback(ctx)
	duration := time.Since(start)

	if err != nil {
		u.logger.Error("failed to rollback transaction", zap.Error(err), zap.Duration("duration", duration))
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to rollback")
	} else {
		u.logger.Debug("transaction rolled back successfully", zap.Duration("duration", duration))
		span.SetStatus(codes.Ok, "success")
	}

	span.SetAttributes(attribute.Float64("duration_ms", duration.Seconds()*1000))
	if u.metrics != nil {
		u.metrics.RecordDatabaseQuery(ctx, "ROLLBACK", "transaction", duration)
	}

	return err
}

func (u *instrumentedUnitOfWork) Execute(ctx context.Context, fn func(ctx context.Context, tx *sql.Tx) error) (err error) {
	start := time.Now()
	ctx, span := u.tracer.StartSpan(ctx, "uow.Execute",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("db.operation", "transaction")),
	)
	defer span.End()

	u.logger.Debug("executing transaction")
	if err := u.Begin(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to begin")
		return err
	}

	committed := false
	rolledBack := false

	defer func() {
		duration := time.Since(start)
		if r := recover(); r != nil {
			_ = u.Rollback(ctx)
			rolledBack = true
			err = fmt.Errorf("panic in transaction: %v", r)
			u.logger.Error("panic in transaction", zap.Any("panic", r), zap.Duration("duration", duration))
			span.RecordError(err, trace.WithAttributes(attribute.String("error.type", "panic")))
			span.SetStatus(codes.Error, "panic")
		}
		if err != nil && !rolledBack {
			_ = u.Rollback(ctx)
			rolledBack = true
			u.logger.Warn("transaction rolled back due to error", zap.Error(err), zap.Duration("duration", duration))
			span.AddEvent("transaction.rollback", trace.WithAttributes(attribute.String("reason", "error")))
		}
		span.SetAttributes(attribute.Float64("duration_ms", duration.Seconds()*1000),
			attribute.Bool("committed", committed),
			attribute.Bool("rolled_back", rolledBack),
		)
	}()

	err = fn(ctx, u.unitOfWork.tx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "execution failed")
		return err
	}

	err = u.Commit(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "commit failed")
		return err
	}

	committed = true
	duration := time.Since(start)
	u.logger.Debug("transaction executed successfully", zap.Duration("duration", duration))
	span.SetStatus(codes.Ok, "success")
	span.AddEvent("transaction.commit")
	if u.metrics != nil {
		u.metrics.RecordDatabaseQuery(ctx, "EXECUTE", "transaction", duration)
	}

	return nil
}

func (u *instrumentedUnitOfWork) IsInTransaction() bool {
	return u.tx != nil
}

func (u *instrumentedUnitOfWork) GetTx() *sql.Tx {
	return u.tx
}
