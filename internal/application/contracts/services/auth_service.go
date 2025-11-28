package services

import (
	"authentication/internal/application/dtos"
	"authentication/internal/domain/aggregates"
	"authentication/internal/domain/valueobjects"
	"context"
)

type AuthService interface {
	Register(ctx context.Context, input dtos.RegisterInput) (*aggregates.UserAggregate, error)
	Login(ctx context.Context, identifier, password, ipAddress, userAgent string) (*dtos.LoginResult, error)
	LoginWithOAuth(ctx context.Context, provider, oauthID, ipAddress, userAgent string) (*dtos.LoginResult, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dtos.TokenPair, error)
	Logout(ctx context.Context, userID, sessionID string) error
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error
	ValidateToken(ctx context.Context, token string) (*valueobjects.TokenClaims, error)
}
