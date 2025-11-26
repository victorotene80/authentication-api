package services

import (
    "context"
    "fmt"
    "authentication/internal/domain/aggregates"
    "authentication/internal/domain/repositories"
    "authentication/internal/domain/valueobjects"
)

type AuthDomainService interface {
    ValidateCredentials(ctx context.Context, identifier, password string) (*aggregates.UserAggregate, error)
    CheckUserExists(ctx context.Context, email valueobjects.Email, username valueobjects.Username) error
}

type authDomainService struct {
    userRepo        repositories.UserRepository
    passwordService PasswordService
}

func NewAuthDomainService(userRepo repositories.UserRepository, passwordService PasswordService) AuthDomainService {
    return &authDomainService{
        userRepo:        userRepo,
        passwordService: passwordService,
    }
}

func (s *authDomainService) ValidateCredentials(ctx context.Context, identifier, password string) (*aggregates.UserAggregate, error) {
    user, err := s.userRepo.FindByEmailOrUsername(ctx, identifier)
    if err != nil {
        if err == repositories.ErrNotFound {
            return nil, fmt.Errorf("invalid credentials")
        }
        return nil, fmt.Errorf("failed to find user: %w", err)
    }

    if !user.User.IsActive {
        return nil, fmt.Errorf("user account is deactivated")
    }

    valid, err := s.passwordService.Verify(password, user.User.PasswordHash)
    if err != nil {
        return nil, fmt.Errorf("failed to verify password: %w", err)
    }

    if !valid {
        return nil, fmt.Errorf("invalid credentials")
    }

    return user, nil
}

func (s *authDomainService) CheckUserExists(ctx context.Context, email valueobjects.Email, username valueobjects.Username) error {
    emailExists, err := s.userRepo.ExistsByEmail(ctx, email)
    if err != nil {
        return fmt.Errorf("failed to check email: %w", err)
    }
    if emailExists {
        return fmt.Errorf("email already exists")
    }

    usernameExists, err := s.userRepo.ExistsByUsername(ctx, username)
    if err != nil {
        return fmt.Errorf("failed to check username: %w", err)
    }
    if usernameExists {
        return fmt.Errorf("username already exists")
    }

    return nil
}
