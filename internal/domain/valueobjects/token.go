package valueobjects

import (
	"fmt"
	"time"
)

type TokenType string

const (
	TokenTypeAccess  TokenType = "ACCESS"
	TokenTypeRefresh TokenType = "REFRESH"
)

type Token struct {
	value     string
	tokenType TokenType
	expiresAt time.Time
	issuedAt  time.Time
}

// NewToken creates a validated Token
func NewToken(value string, tokenType TokenType, expiresAt time.Time) (*Token, error) {
	if value == "" {
		return nil, fmt.Errorf("token value cannot be empty")
	}
	if !tokenType.IsValid() {
		return nil, fmt.Errorf("invalid token type: %s", tokenType)
	}
	if expiresAt.Before(time.Now().UTC()) {
		return nil, fmt.Errorf("expiration time must be in the future")
	}

	return &Token{
		value:     value,
		tokenType: tokenType,
		expiresAt: expiresAt,
		issuedAt:  time.Now().UTC(),
	}, nil
}

func (t TokenType) IsValid() bool {
	return t == TokenTypeAccess || t == TokenTypeRefresh
}

func (t *Token) Value() string {
	return t.value
}

func (t *Token) Type() TokenType {
	return t.tokenType
}

func (t *Token) IssuedAt() time.Time {
	return t.issuedAt
}

func (t *Token) ExpiresAt() time.Time {
	return t.expiresAt
}

func (t *Token) RemainingValidity() time.Duration {
	return t.expiresAt.Sub(time.Now().UTC())
}

func (t *Token) IsExpired() bool {
	return time.Now().UTC().After(t.expiresAt)
}

func (t *Token) IsValid() bool {
	return !t.IsExpired() && t.value != ""
}
