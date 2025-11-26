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
    Value     string
    Type      TokenType
    ExpiresAt time.Time
    IssuedAt  time.Time
}

func NewToken(value string, tokenType TokenType, expiresAt time.Time) (*Token, error) {
    if value == "" {
        return nil, fmt.Errorf("token value cannot be empty")
    }
    if !tokenType.IsValid() {
        return nil, fmt.Errorf("invalid token type: %s", tokenType)
    }
    return &Token{
        Value:     value,
        Type:      tokenType,
        ExpiresAt: expiresAt,
        IssuedAt:  time.Now(),
    }, nil
}

func (t TokenType) IsValid() bool {
    return t == TokenTypeAccess || t == TokenTypeRefresh
}

func (t *Token) IsExpired() bool {
    return time.Now().After(t.ExpiresAt)
}

func (t *Token) IsValid() bool {
    return !t.IsExpired() && t.Value != ""
}