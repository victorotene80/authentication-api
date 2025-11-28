package valueobjects

type TokenClaims struct {
	UserID    string `json:"uid"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	SessionID string `json:"sid"`
	TokenType string `json:"typ"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
	JTI       string `json:"jti"`
}