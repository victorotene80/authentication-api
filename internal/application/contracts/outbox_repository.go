package contracts

import (
    "context"
    "github.com/google/uuid"
	"authentication/internal/domain/entities"
	uow "authentication/shared/persistence"
)


type OutboxRepository interface {
	InsertTx(ctx context.Context, db uow.DB, event entities.OutboxEvent) error
    FetchPending(ctx context.Context, limit int) ([]entities.OutboxEvent, error)
    MarkSent(ctx context.Context, id uuid.UUID) error
    MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error
	SetDB(db uow.DB)

}
