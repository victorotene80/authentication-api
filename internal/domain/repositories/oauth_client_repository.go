package repositories

import (
	"context"
	"authentication/internal/domain/aggregates"
	uow "authentication/internal/application/contracts/persistence"
)

type OAuthClientRepository interface {
	Create(ctx context.Context, tx uow.DB, client *aggregates.OAuthClient) error
	FindByID(ctx context.Context, tx uow.DB, id string) (*aggregates.OAuthClient, error)
	FindByClientID(ctx context.Context, tx uow.DB, clientID string) (*aggregates.OAuthClient, error)
	FindByProvider(ctx context.Context, tx uow.DB, provider string) ([]*aggregates.OAuthClient, error)
	Update(ctx context.Context, tx uow.DB, client *aggregates.OAuthClient) error
	Delete(ctx context.Context, tx uow.DB, id string) error
	List(ctx context.Context, tx uow.DB, page, pageSize int) ([]*aggregates.OAuthClient, int64, error)
}