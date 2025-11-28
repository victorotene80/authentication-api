package repositories

import (
	"context"
	"authentication/internal/domain/entities"
)

type OAuthTokenRepository interface {
	Create(ctx context.Context, token *entities.OAuthToken) error
	FindByID(ctx context.Context, id string) (*entities.OAuthToken, error)
	FindByUserIDAndProvider(ctx context.Context, userID, provider string) (*entities.OAuthToken, error)
	FindByUserID(ctx context.Context, userID string) ([]*entities.OAuthToken, error)
	Update(ctx context.Context, token *entities.OAuthToken) error
	Delete(ctx context.Context, id string) error
	DeleteByUserID(ctx context.Context, userID string) error
	RevokeByUserIDAndProvider(ctx context.Context, userID, provider string) error
}