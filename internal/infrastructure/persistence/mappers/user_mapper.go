// internal/infrastructure/persistence/mappers/user_mapper.go
package mappers

import (
    "authentication/internal/domain/entities"
    "authentication/internal/domain/valueobjects"
    "authentication/internal/infrastructure/persistence/database/models"
)

type UserMapper struct{}

func NewUserMapper() *UserMapper {
    return &UserMapper{}
}

// ToModel converts domain entity to database model
func (m *UserMapper) ToModel(user *entities.User) *models.UserModel {
    var phone *string
    if user.Phone().String() != "" {
        p := user.Phone().String()
        phone = &p
    }
    
    var twoFactorSecret *string
    if user.TwoFactorSecret() != "" {
        secret := user.TwoFactorSecret()
        twoFactorSecret = &secret
    }

    return &models.UserModel{
        ID:                  user.ID(),
        Email:               user.Email().String(),
        Username:            user.Username().String(),
        PasswordHash:        user.PasswordHash(),
        FirstName:           user.FirstName(),
        LastName:            user.LastName(),
        Phone:               phone,
        EmailVerified:       user.IsEmailVerified(),
        EmailVerifiedAt:     user.EmailVerifiedAt(),
        Status:              string(user.Status()),
        IsLocked:            user.IsLocked(),
        LockedUntil:         user.LockedUntil(),
        FailedLoginAttempts: user.FailedLoginAttempts(),
        LastFailedLogin:     user.LastFailedLogin(),
        TwoFactorEnabled:    user.TwoFactorEnabled(),
        TwoFactorSecret:     twoFactorSecret,
        CreatedAt:           user.CreatedAt(),
        UpdatedAt:           user.UpdatedAt(),
        LastLogin:           user.LastLogin(),
    }
}



func (m *UserMapper) ToDomain(model *models.UserModel) (*entities.User, error) {
    email, err := valueobjects.NewEmail(model.Email)
    if err != nil {
        return nil, err
    }

    username, err := valueobjects.NewUsername(model.Username)
    if err != nil {
        return nil, err
    }

	//comeback to this

    password, err := valueobjects.NewPasswordFromHash(model.PasswordHash)
    if err != nil {
        return nil, err
    }
    var phone valueobjects.PhoneNumber
    if model.Phone != nil {
        phone, err = valueobjects.NewPhoneNumber(*model.Phone)
        if err != nil {
            return nil, err
        }
    }

    return entities.ReconstructUser(
        model.ID,
        email,
        username,
        password,
        model.FirstName,
        model.LastName,
        phone,
        model.EmailVerified,
        model.EmailVerifiedAt,
        entities.UserStatus(model.Status),
        model.IsLocked,
        model.LockedUntil,
        model.FailedLoginAttempts,
        model.LastFailedLogin,
        model.TwoFactorEnabled,
        model.TwoFactorSecret,
        model.CreatedAt,
        model.UpdatedAt,
        model.LastLogin,
    ), nil
}