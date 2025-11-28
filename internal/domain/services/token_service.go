package services

import "authentication/internal/domain/valueobjects"


type TokenService interface {
	GenerateAccessToken(userID, email, username, role, sessionID string) (*valueobjects.Token, error)
	GenerateRefreshToken(userID, sessionID string) (*valueobjects.Token, error)
	ValidateToken(token string) (*valueobjects.TokenClaims, error)
	ExtractClaims(token string) (*valueobjects.TokenClaims, error)
}