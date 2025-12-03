package repositories

import (
	"context"
	"authentication/internal/domain/aggregates"
	"authentication/internal/domain/valueobjects"
)

type UserRepository interface {
	Create(ctx context.Context, user *aggregates.UserAggregate) error
	FindByID(ctx context.Context, id string) (*aggregates.UserAggregate, error)
	FindByEmail(ctx context.Context, email valueobjects.Email) (*aggregates.UserAggregate, error)
	FindByUsername(ctx context.Context, username valueobjects.Username) (*aggregates.UserAggregate, error)
	ExistsByEmail(ctx context.Context, email valueobjects.Email) (bool, error)
	ExistsByUsername(ctx context.Context, username valueobjects.Username) (bool, error)
	Update(ctx context.Context, user *aggregates.UserAggregate) error
	Delete(ctx context.Context, id string) error
	List(
		ctx context.Context,
		page, pageSize int,
		role *valueobjects.Role,
		isActive *bool,
	) ([]*aggregates.UserAggregate, int64, error)
}