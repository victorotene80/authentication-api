package entities

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	RefreshTokenHash string
	TokenFamily      uuid.UUID
	DeviceID         string
	DeviceName       string
	UserAgent        string
	IPAddress        string
	CountryCode      string
	City             string
	CreatedAt        time.Time
	LastActivityAt   time.Time
	ExpiresAt        time.Time
	IsActive         bool
	EndedAt          *time.Time
	EndReason        string
}

func NewSession(
	userID uuid.UUID,
	refreshTokenHash string,
	deviceID, deviceName, userAgent, ipAddress string,
	expiresAt time.Time,
) *Session {
	now := time.Now()
	return &Session{
		ID:               uuid.New(),
		UserID:           userID,
		RefreshTokenHash: refreshTokenHash,
		TokenFamily:      uuid.New(),
		DeviceID:         deviceID,
		DeviceName:       deviceName,
		UserAgent:        userAgent,
		IPAddress:        ipAddress,
		CreatedAt:        now,
		LastActivityAt:   now,
		ExpiresAt:        expiresAt,
		IsActive:         true,
	}
}

func (s *Session) UpdateActivity() {
	s.LastActivityAt = time.Now()
}

func (s *Session) End(reason string) {
	now := time.Now()
	s.IsActive = false
	s.EndedAt = &now
	s.EndReason = reason
}

func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

func (s *Session) IsValid() bool {
	return s.IsActive && !s.IsExpired()
}
