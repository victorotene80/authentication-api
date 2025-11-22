package uow

import (
	uow "authentication/shared/persistence"
	"context"
	"database/sql"
	"fmt"
)

type UnitOfWork struct {
	db *sql.DB
	tx *sql.Tx
}

func NewUnitOfWork(db *sql.DB) uow.UnitOfWork {
	return &UnitOfWork{db: db}
}

func (u *UnitOfWork) Begin(ctx context.Context) error {
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

func (u *UnitOfWork) Commit(ctx context.Context) error {
	if u.tx == nil {
		return fmt.Errorf("no active transaction")
	}

	if err := u.tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	u.tx = nil
	return nil
}

func (u *UnitOfWork) Rollback(ctx context.Context) error {
	if u.tx == nil {
		return nil
	}

	if err := u.tx.Rollback(); err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	u.tx = nil
	return nil
}

func (u *UnitOfWork) IsInTransaction() bool {
	return u.tx != nil
}

// NEW SIGNATURE — recommended best practice
func (u *UnitOfWork) Execute(
	ctx context.Context,
	fn func(ctx context.Context, tx *sql.Tx) error,
) (err error) {

	if err := u.Begin(ctx); err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			_ = u.Rollback(ctx)
			err = fmt.Errorf("panic: %v", r)
			return
		}

		if err != nil {
			_ = u.Rollback(ctx)
			return
		}
	}()

	err = fn(ctx, u.tx)
	if err != nil {
		return err
	}

	return u.Commit(ctx)
}
