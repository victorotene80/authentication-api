package repositories

import (
	"context"

	"authentication/internal/domain/entities"
)

type SessionRepository interface {
	Create(ctx context.Context, session *entities.Session) error
	FindByID(ctx context.Context, id string) (*entities.Session, error)
	FindByRefreshToken(ctx context.Context, refreshToken string) (*entities.Session, error)
	FindByUserID(ctx context.Context, userID string) ([]*entities.Session, error)
	FindActiveByUserID(ctx context.Context, userID string) ([]*entities.Session, error)
	Update(ctx context.Context, session *entities.Session) error
	Delete(ctx context.Context, id string) error
	DeleteByUserID(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
	RevokeByUserID(ctx context.Context, userID string) error
	RevokeByID(ctx context.Context, id string) error
}
