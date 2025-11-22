package repositories

import (
	"context"
	"time"
	"github.com/google/uuid"
	"authentication/internal/domain/entities"
)

type SessionRepository interface {
	// Session operations
	CreateSession(ctx context.Context, session *entities.Session) error
	GetSessionByID(ctx context.Context, id uuid.UUID) (*entities.Session, error)
	GetActiveSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Session, error)
	UpdateSession(ctx context.Context, session *entities.Session) error
	DeleteSession(ctx context.Context, id uuid.UUID) error
	EndSession(ctx context.Context, id uuid.UUID, reason string) error
	EndAllUserSessions(ctx context.Context, userID uuid.UUID, reason string) error
	
	// Refresh token operations
	CreateRefreshToken(ctx context.Context, token *entities.RefreshToken) error
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*entities.RefreshToken, error)
	GetRefreshTokensByFamily(ctx context.Context, tokenFamily uuid.UUID) ([]*entities.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string, reason string) error
	RevokeTokenFamily(ctx context.Context, tokenFamily uuid.UUID, reason string) error
	RevokeAllUserTokens(ctx context.Context, userID uuid.UUID, reason string) error
	DeleteExpiredTokens(ctx context.Context) error
	
	// Token blacklist operations
	BlacklistToken(ctx context.Context, jti string, userID uuid.UUID, expiresAt time.Time, reason string) error
	IsTokenBlacklisted(ctx context.Context, jti string) (bool, error)
	CleanupExpiredBlacklist(ctx context.Context) error
}