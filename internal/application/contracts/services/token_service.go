package services

import (
	"context"
	"time"
)

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	ExpiresIn    int64     `json:"expires_in"` 
	TokenType    string    `json:"token_type"` 
	SessionID    string    `json:"session_id"`
}

type TokenClaims struct {
	UserID    string
	TokenID   string
	SessionID string
	Role      string
	Email     string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

type SessionMetadata struct {
	IPAddress string
	UserAgent string
	DeviceID  string
}

type TokenService interface {
	Generate(ctx context.Context, userID, role, email string, metadata SessionMetadata) (*TokenPair, error)
	VerifyAccess(ctx context.Context, token string) (*TokenClaims, error)
	VerifyRefresh(ctx context.Context, refreshToken string) (*TokenClaims, error)
	RefreshTokens(ctx context.Context, refreshToken string, metadata SessionMetadata) (*TokenPair, error)
	RevokeSession(ctx context.Context, sessionID string) error
	RevokeAllUserSessions(ctx context.Context, userID string) error
	IsSessionValid(ctx context.Context, sessionID string) (bool, error)
	GetActiveSessions(ctx context.Context, userID string) ([]SessionInfo, error)
}

type SessionInfo struct {
	SessionID string
	UserID    string
	CreatedAt time.Time
	ExpiresAt time.Time
	IPAddress string
	UserAgent string
	DeviceID  string
	LastUsed  time.Time
}