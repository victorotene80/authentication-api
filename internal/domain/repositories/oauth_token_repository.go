package repositories

import (
	"context"
	"authentication/internal/domain/entities"
	uow "authentication/internal/application/contracts/persistence"
)

type OAuthTokenRepository interface {
	Create(ctx context.Context, tx uow.DB, token *entities.OAuthToken) error
	FindByID(ctx context.Context, tx uow.DB, id string) (*entities.OAuthToken, error)
	FindByUserIDAndProvider(ctx context.Context, tx uow.DB, userID, provider string) (*entities.OAuthToken, error)
	FindByUserID(ctx context.Context, tx uow.DB, userID string) ([]*entities.OAuthToken, error)
	Update(ctx context.Context, tx uow.DB, token *entities.OAuthToken) error
	Delete(ctx context.Context, tx uow.DB, id string) error
	DeleteByUserID(ctx context.Context, tx uow.DB, userID string) error
	RevokeByUserIDAndProvider(ctx context.Context, tx uow.DB, userID, provider string) error
}