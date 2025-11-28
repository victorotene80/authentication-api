package infrastructure

import (
	"fmt"
	"time"

	"authentication/internal/domain/services"
	"authentication/internal/domain/valueobjects"

	"github.com/golang-jwt/jwt/v5"
)

type JWTTokenService struct {
	accessSecret      []byte
	refreshSecret     []byte
	accessExpiration  time.Duration
	refreshExpiration time.Duration
	issuer            string
}

func NewJWTTokenService(accessSecret, refreshSecret string, accessExp, refreshExp time.Duration, issuer string) *JWTTokenService {
	return &JWTTokenService{
		accessSecret:      []byte(accessSecret),
		refreshSecret:     []byte(refreshSecret),
		accessExpiration:  accessExp,
		refreshExpiration: refreshExp,
		issuer:            issuer,
	}
}

type jwtClaims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateAccessToken creates a signed access token
func (s *JWTTokenService) GenerateAccessToken(userID, email, username, role string) (*valueobjects.Token, error) {
	now := time.Now().UTC()
	expiration := now.Add(s.accessExpiration)

	claims := &jwtClaims{
		UserID:   userID,
		Email:    email,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.issuer,
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.accessSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	return valueobjects.NewToken(signed, valueobjects.TokenTypeAccess, expiration)
}

// GenerateRefreshToken creates a signed refresh token
func (s *JWTTokenService) GenerateRefreshToken(userID string) (*valueobjects.Token, error) {
	now := time.Now().UTC()
	expiration := now.Add(s.refreshExpiration)

	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expiration),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		Issuer:    s.issuer,
		Subject:   userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.refreshSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return valueobjects.NewToken(signed, valueobjects.TokenTypeRefresh, expiration)
}

// ValidateToken validates an access token
func (s *JWTTokenService) ValidateToken(tokenString string) (*services.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.accessSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("token parse error: %w", err)
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return &services.TokenClaims{
		UserID:   claims.UserID,
		Email:    claims.Email,
		Username: claims.Username,
		Role:     claims.Role,
	}, nil
}

// ExtractClaims delegates to ValidateToken
func (s *JWTTokenService) ExtractClaims(token string) (*services.TokenClaims, error) {
	return s.ValidateToken(token)
}
