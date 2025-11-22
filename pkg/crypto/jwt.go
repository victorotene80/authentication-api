package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims represents JWT claims
type Claims struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Roles     []string  `json:"roles"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"refresh_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	TokenType             string    `json:"token_type"`
}

// JWTService handles JWT token operations
type JWTService struct {
	accessSecret         []byte
	refreshSecret        []byte
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
	issuer               string
}

// NewJWTServiceFromEnv creates a new JWT service using environment variables
func NewJWTServiceFromEnv() *JWTService {
	accessSecret := os.Getenv("JWT_ACCESS_SECRET")
	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	issuer := os.Getenv("JWT_ISSUER")

	// Fail fast if secrets are missing
	if accessSecret == "" || refreshSecret == "" {
		panic("JWT_ACCESS_SECRET and JWT_REFRESH_SECRET must be set in environment")
	}

	// Default durations if not provided
	accessTokenDuration := parseDurationOrDefault(os.Getenv("JWT_ACCESS_TOKEN_DURATION"), 15*time.Minute)
	refreshTokenDuration := parseDurationOrDefault(os.Getenv("JWT_REFRESH_TOKEN_DURATION"), 7*24*time.Hour)

	return &JWTService{
		accessSecret:         []byte(accessSecret),
		refreshSecret:        []byte(refreshSecret),
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
		issuer:               issuerOrDefault(issuer),
	}
}

// parseDurationOrDefault safely parses duration strings (e.g. "15m", "1h")
func parseDurationOrDefault(value string, def time.Duration) time.Duration {
	if value == "" {
		return def
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return def
	}
	return d
}

// issuerOrDefault sets a fallback for missing issuer
func issuerOrDefault(issuer string) string {
	if issuer == "" {
		hostname, _ := os.Hostname()
		return fmt.Sprintf("auth-service@%s", hostname)
	}
	return issuer
}

// GenerateTokenPair generates both access and refresh tokens
func (s *JWTService) GenerateTokenPair(userID, email string, roles []string) (*TokenPair, error) {
	now := time.Now()

	accessToken, accessExp, err := s.generateToken(
		userID, email, roles, AccessToken, now, s.accessTokenDuration, s.accessSecret,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, refreshExp, err := s.generateToken(
		userID, email, roles, RefreshToken, now, s.refreshTokenDuration, s.refreshSecret,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessExp,
		RefreshTokenExpiresAt: refreshExp,
		TokenType:             "Bearer",
	}, nil
}

// generateToken creates a JWT token
func (s *JWTService) generateToken(
	userID, email string,
	roles []string,
	tokenType TokenType,
	now time.Time,
	duration time.Duration,
	secret []byte,
) (string, time.Time, error) {
	expiresAt := now.Add(duration)

	claims := Claims{
		UserID:    userID,
		Email:     email,
		Roles:     roles,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Subject:   userID,
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return signedToken, expiresAt, nil
}

// ValidateAccessToken validates an access token
func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	return s.validateToken(tokenString, s.accessSecret, AccessToken)
}

// ValidateRefreshToken validates a refresh token
func (s *JWTService) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return s.validateToken(tokenString, s.refreshSecret, RefreshToken)
}

// validateToken validates a token with the given secret
func (s *JWTService) validateToken(tokenString string, secret []byte, expectedType TokenType) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrTokenNotYetValid
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("invalid token type: expected %s, got %s", expectedType, claims.TokenType)
	}

	return claims, nil
}

// GenerateRandomToken generates a cryptographically secure random token
func GenerateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

/*// HashToken creates a hash of a token for storage
func HashToken(token string) string {
	return HashPassword(token) // reuse bcrypt or your existing hash method
}

func HashPassword(token string) string {
	panic("unimplemented")
}*/
