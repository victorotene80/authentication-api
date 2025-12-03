package services

import (
	"context"
	"time"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

type TokenClaims struct {
	UserID   string
	Email    string
	Username string
	Role     string
}

type TokenService interface {
	Generate(ctx context.Context, claims TokenClaims) (*TokenPair, error)
	Verify(ctx context.Context, accessToken string) (*TokenClaims, error)
	Refresh(ctx context.Context, refreshToken string) (*TokenPair, error)
	Revoke(ctx context.Context, refreshToken string) error
}
