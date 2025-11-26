package repositories

import (
	"context"

	uow "authentication/internal/application/contracts/persistence"
	"authentication/internal/domain/entities"
)

type SessionRepository interface {
	Create(ctx context.Context, tx uow.DB, session *entities.Session) error
	FindByID(ctx context.Context, tx uow.DB, id string) (*entities.Session, error)
	FindByRefreshToken(ctx context.Context, tx uow.DB, refreshToken string) (*entities.Session, error)
	FindByUserID(ctx context.Context, tx uow.DB, userID string) ([]*entities.Session, error)
	FindActiveByUserID(ctx context.Context, tx uow.DB, userID string) ([]*entities.Session, error)
	Update(ctx context.Context, tx uow.DB, session *entities.Session) error
	Delete(ctx context.Context, tx uow.DB, id string) error
	DeleteByUserID(ctx context.Context, tx uow.DB, userID string) error
	DeleteExpired(ctx context.Context, tx uow.DB) error
	RevokeByUserID(ctx context.Context, tx uow.DB, userID string) error
	RevokeByID(ctx context.Context, tx uow.DB, id string) error
}
