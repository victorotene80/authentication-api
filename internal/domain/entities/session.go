package entities

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID           string
	UserID       string
	RefreshToken string
	AccessToken  string
	IPAddress    string
	UserAgent    string
	ExpiresAt    time.Time
	IsRevoked    bool
	RevokedAt    *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewSession(
	userID string,
	refreshToken string,
	accessToken string,
	ipAddress string,
	userAgent string,
	expiresAt time.Time,
) *Session {
	now := time.Now()
	return &Session{
		ID:           uuid.New().String(),
		UserID:       userID,
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		ExpiresAt:    expiresAt,
		IsRevoked:    false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (s *Session) Revoke() {
	now := time.Now()
	s.IsRevoked = true
	s.RevokedAt = &now
	s.UpdatedAt = now
}

func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

func (s *Session) IsValid() bool {
	return !s.IsRevoked && !s.IsExpired()
}

func (s *Session) UpdateTokens(refreshToken, accessToken string, expiresAt time.Time) {
	s.RefreshToken = refreshToken
	s.AccessToken = accessToken
	s.ExpiresAt = expiresAt
	s.UpdatedAt = time.Now()
}
