package entities

import (
	"time"

	"github.com/google/uuid"
)

type OAuthToken struct {
	ID           string
	UserID       string
	Provider     string
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresAt    time.Time
	Scope        string
	IsRevoked    bool
	RevokedAt    *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewOAuthToken(
	userID string,
	provider string,
	accessToken string,
	refreshToken string,
	tokenType string,
	expiresAt time.Time,
	scope string,
) *OAuthToken {
	now := time.Now()
	return &OAuthToken{
		ID:           uuid.New().String(),
		UserID:       userID,
		Provider:     provider,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    tokenType,
		ExpiresAt:    expiresAt,
		Scope:        scope,
		IsRevoked:    false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (o *OAuthToken) Revoke() {
	now := time.Now()
	o.IsRevoked = true
	o.RevokedAt = &now
	o.UpdatedAt = now
}

func (o *OAuthToken) IsExpired() bool {
	return time.Now().After(o.ExpiresAt)
}

func (o *OAuthToken) IsValid() bool {
	return !o.IsRevoked && !o.IsExpired()
}

func (o *OAuthToken) Update(accessToken, refreshToken string, expiresAt time.Time) {
	o.AccessToken = accessToken
	if refreshToken != "" {
		o.RefreshToken = refreshToken
	}
	o.ExpiresAt = expiresAt
	o.UpdatedAt = time.Now()
}
