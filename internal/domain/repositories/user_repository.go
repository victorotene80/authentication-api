package repositories

import (
	"context"

	"authentication/internal/domain/entities"
	"authentication/internal/domain/valueobjects"
	uow "authentication/shared/persistence"

	"github.com/google/uuid"
)

type UserRepository interface {
	Save(ctx context.Context, tx uow.DB, user *entities.User) error
	FindByID(ctx context.Context, tx uow.DB, id uuid.UUID) (*entities.User, error)
	FindByEmail(ctx context.Context, tx uow.DB, email valueobjects.Email) (*entities.User, error)
	FindByUsername(ctx context.Context, tx uow.DB, username valueobjects.Username) (*entities.User, error)
	ExistsByEmail(ctx context.Context, tx uow.DB, email valueobjects.Email) (bool, error)
	ExistsByUsername(ctx context.Context, tx uow.DB, username valueobjects.Username) (bool, error)
}