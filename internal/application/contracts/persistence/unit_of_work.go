package persistence

import (
    "context"
    "database/sql"
)


type UnitOfWork interface {
	Begin(ctx context.Context) error
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	IsInTransaction() bool
	Con() DB

	Execute(ctx context.Context, fn func(ctx context.Context) error) error
	GetTx() *sql.Tx
}

