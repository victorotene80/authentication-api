package responses

import "time"

type SessionResponse struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	IPAddress      string     `json:"ip_address"`
	UserAgent      string     `json:"user_agent"`
	DeviceType     string     `json:"device_type"`
	Location       *Location  `json:"location,omitempty"`
	IsCurrent      bool       `json:"is_current"`
	IsActive       bool       `json:"is_active"`
	IsSuspicious   bool       `json:"is_suspicious,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	ExpiresAt      time.Time  `json:"expires_at"`
	LastActivityAt time.Time  `json:"last_activity_at"`
}

type Location struct {
	Country string `json:"country,omitempty"`
	City    string `json:"city,omitempty"`
	Region  string `json:"region,omitempty"`
}