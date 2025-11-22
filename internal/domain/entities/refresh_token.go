package entities
import (
	"time"
	
	"github.com/google/uuid"
)

type RefreshToken struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	TokenHash   string
	TokenFamily uuid.UUID
	DeviceID    string
	DeviceName  string
	UserAgent   string
	IPAddress   string
	ExpiresAt   time.Time
	CreatedAt   time.Time
	LastUsedAt  *time.Time
	IsRevoked   bool
	RevokedAt   *time.Time
	RevokedReason string
}

func NewRefreshToken(
	userID uuid.UUID,
	tokenHash string,
	tokenFamily uuid.UUID,
	deviceID, deviceName, userAgent, ipAddress string,
	expiresAt time.Time,
) *RefreshToken {
	return &RefreshToken{
		ID:          uuid.New(),
		UserID:      userID,
		TokenHash:   tokenHash,
		TokenFamily: tokenFamily,
		DeviceID:    deviceID,
		DeviceName:  deviceName,
		UserAgent:   userAgent,
		IPAddress:   ipAddress,
		ExpiresAt:   expiresAt,
		CreatedAt:   time.Now(),
		IsRevoked:   false,
	}
}

func (rt *RefreshToken) MarkAsUsed() {
	now := time.Now()
	rt.LastUsedAt = &now
}

func (rt *RefreshToken) Revoke(reason string) {
	now := time.Now()
	rt.IsRevoked = true
	rt.RevokedAt = &now
	rt.RevokedReason = reason
}

func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

func (rt *RefreshToken) IsValid() bool {
	return !rt.IsRevoked && !rt.IsExpired()
}