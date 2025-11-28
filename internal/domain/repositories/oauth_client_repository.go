package repositories

import (
	"context"
	"authentication/internal/domain/aggregates"
)

type OAuthClientRepository interface {
	Create(ctx context.Context, client *aggregates.OAuthClient) error
	FindByID(ctx context.Context, id string) (*aggregates.OAuthClient, error)
	FindByClientID(ctx context.Context, clientID string) (*aggregates.OAuthClient, error)
	FindByProvider(ctx context.Context, provider string) ([]*aggregates.OAuthClient, error)
	Update(ctx context.Context, client *aggregates.OAuthClient) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, page, pageSize int) ([]*aggregates.OAuthClient, int64, error)
}