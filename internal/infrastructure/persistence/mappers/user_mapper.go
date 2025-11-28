package mappers

import (
	"authentication/internal/domain/aggregates"
	"authentication/internal/domain/entities"
	"authentication/internal/domain/valueobjects"
	"authentication/internal/infrastructure/persistence/database/models"
)

type UserMapper struct{}

func NewUserMapper() *UserMapper {
	return &UserMapper{}
}

func (m *UserMapper) ToModel(aggregate *aggregates.UserAggregate) *models.UserModel {
	return &models.UserModel{
		ID:           aggregate.User.ID,
		Username:     aggregate.User.Username.String(),
		Email:        aggregate.User.Email.String(),
		PasswordHash: aggregate.User.Password.Value(),
		Phone:        aggregate.User.Phone.String(),
		FirstName:    aggregate.User.FirstName,
		LastName:     aggregate.User.LastName,
		Role:         aggregate.User.Role.String(),
		IsActive:     aggregate.User.IsActive,
		IsVerified:   aggregate.User.IsVerified,
		LastLoginAt:  aggregate.User.LastLoginAt,
		Version:      aggregate.Version(),
		CreatedAt:    aggregate.User.CreatedAt,
		UpdatedAt:    aggregate.User.UpdatedAt,
	}
}

func (m *UserMapper) ToDomain(model *models.UserModel) (*aggregates.UserAggregate, error) {
	username, err := valueobjects.NewUsername(model.Username)
	if err != nil {
		return nil, err
	}

	email, err := valueobjects.NewEmail(model.Email)
	if err != nil {
		return nil, err
	}

	phone, err := valueobjects.NewPhoneNumber(model.Phone)
	if err != nil {
		return nil, err
	}

	role, err := valueobjects.NewRole(model.Role)
	if err != nil {
		return nil, err
	}

	user := &entities.User{
		ID:          model.ID,
		Username:    username,
		Email:       email,
		Password:    valueobjects.NewPassword(model.PasswordHash),
		Phone:       phone,
		FirstName:   model.FirstName,
		LastName:    model.LastName,
		Role:        role,
		IsActive:    model.IsActive,
		IsVerified:  model.IsVerified,
		LastLoginAt: model.LastLoginAt,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}

	aggregate := &aggregates.UserAggregate{
		AggregateRoot: aggregates.NewAggregateRoot(model.ID),
		User:          user,
		Sessions:      make([]*entities.Session, 0),
	}

	return aggregate, nil
}
