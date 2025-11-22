package uow

import (
	"authentication/internal/infrastructure/observability/metrics"
	uow "authentication/shared/persistence"
	"context"
	"time"
	"database/sql"
)

type InstrumentedUnitOfWork struct {
	wrapped uow.UnitOfWork
}

func NewInstrumentedUnitOfWork(base uow.UnitOfWork) uow.UnitOfWork {
	return &InstrumentedUnitOfWork{wrapped: base}
}

func (i *InstrumentedUnitOfWork) Begin(ctx context.Context) error {
	return i.wrapped.Begin(ctx)
}

func (i *InstrumentedUnitOfWork) Commit(ctx context.Context) error {
	return i.wrapped.Commit(ctx)
}

func (i *InstrumentedUnitOfWork) Rollback(ctx context.Context) error {
	return i.wrapped.Rollback(ctx)
}

func (i *InstrumentedUnitOfWork) IsInTransaction() bool {
	return i.wrapped.IsInTransaction()
}

func (i *InstrumentedUnitOfWork) Execute(
	ctx context.Context,
	fn func(ctx context.Context, tx *sql.Tx) error,
) error {

	start := time.Now()
	err := i.wrapped.Execute(ctx, fn)

	duration := time.Since(start).Seconds()
	metrics.DatabaseTransactionDuration.Observe(duration)

	status := "committed"
	if err != nil {
		status = "rolled_back"
	}

	metrics.DatabaseTransactionsTotal.WithLabelValues(status).Inc()

	return err
}
