package repositories

import (
	"context"

	uow "authentication/internal/application/contracts/persistence"
	"authentication/internal/domain/aggregates"
	"authentication/internal/domain/valueobjects"
)

type UserRepository interface {
	Create(ctx context.Context, tx uow.DB, user *aggregates.UserAggregate) error
	FindByID(ctx context.Context, tx uow.DB, id string) (*aggregates.UserAggregate, error)
	FindByEmail(ctx context.Context, tx uow.DB, email valueobjects.Email) (*aggregates.UserAggregate, error)
	FindByUsername(ctx context.Context, tx uow.DB, username valueobjects.Username) (*aggregates.UserAggregate, error)
	FindByEmailOrUsername(ctx context.Context, tx uow.DB, identifier string) (*aggregates.UserAggregate, error)
	Update(ctx context.Context, tx uow.DB, user *aggregates.UserAggregate) error
	Delete(ctx context.Context, tx uow.DB, id string) error
	List(
		ctx context.Context,
		tx uow.DB,
		page, pageSize int,
		role *valueobjects.Role,
		isActive *bool,
	) ([]*aggregates.UserAggregate, int64, error)
	ExistsByEmail(ctx context.Context, tx uow.DB, email valueobjects.Email) (bool, error)
	ExistsByUsername(ctx context.Context, tx uow.DB, username valueobjects.Username) (bool, error)
}
