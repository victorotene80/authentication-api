package uow

import (
	"context"
	"database/sql"
	"fmt"

	uow "authentication/internal/application/contracts/persistence"
)

// unitOfWork is the base implementation without observability
type unitOfWork struct {
	db *sql.DB
	tx *sql.Tx
}

// NewUnitOfWork creates a new unit of work
// Note: For production use, wrap with NewInstrumentedUnitOfWork for observability
func NewUnitOfWork(db *sql.DB) uow.UnitOfWork {
	return &unitOfWork{
		db: db,
	}
}

// Begin starts a new transaction
func (u *unitOfWork) Begin(ctx context.Context) error {
	if u.tx != nil {
		return fmt.Errorf("transaction already started")
	}

	tx, err := u.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	u.tx = tx
	return nil
}

// Commit commits the current transaction
func (u *unitOfWork) Commit(ctx context.Context) error {
	if u.tx == nil {
		return fmt.Errorf("no active transaction")
	}

	if err := u.tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	u.tx = nil
	return nil
}

// Rollback rolls back the current transaction
func (u *unitOfWork) Rollback(ctx context.Context) error {
	if u.tx == nil {
		return nil // No active transaction, nothing to rollback
	}

	if err := u.tx.Rollback(); err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	u.tx = nil
	return nil
}

// IsInTransaction returns true if a transaction is active
func (u *unitOfWork) IsInTransaction() bool {
	return u.tx != nil
}

// GetTx returns the current transaction
func (u *unitOfWork) GetTx() *sql.Tx {
	return u.tx
}

// Execute runs a function within a transaction with automatic commit/rollback
func (u *unitOfWork) Execute(
	ctx context.Context,
	fn func(ctx context.Context, tx *sql.Tx) error,
) (err error) {
	// Begin transaction
	if err := u.Begin(ctx); err != nil {
		return err
	}

	// Defer rollback on panic or error
	defer func() {
		if r := recover(); r != nil {
			_ = u.Rollback(ctx)
			err = fmt.Errorf("panic in transaction: %v", r)
			return
		}

		if err != nil {
			_ = u.Rollback(ctx)
			return
		}
	}()

	// Execute function
	err = fn(ctx, u.tx)
	if err != nil {
		return err
	}

	// Commit transaction
	return u.Commit(ctx)
}