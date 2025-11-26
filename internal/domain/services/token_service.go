package services

import (
    "time"
)

type TokenClaims struct {
    UserID   string `json:"user_id"`
    Email    string `json:"email"`
    Username string `json:"username"`
    Role     string `json:"role"`
    // JWT RegisteredClaims can be added in the infrastructure implementation
}

type TokenService interface {
    GenerateAccessToken(userID, email, username, role string) (string, time.Time, error)
    GenerateRefreshToken(userID string) (string, time.Time, error)
    ValidateToken(token string) (*TokenClaims, error)
    ExtractClaims(token string) (*TokenClaims, error)
}
