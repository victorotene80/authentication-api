package admin

import "time"

type AdminUserResponse struct {
	ID                  string     `json:"id"`
	Email               string     `json:"email"`
	Username            string     `json:"username"`
	FullName            string     `json:"full_name"`
	Status              string     `json:"status"`
	Role                string     `json:"role"`
	EmailVerified       bool       `json:"email_verified"`
	TwoFactorEnabled    bool       `json:"two_factor_enabled"`
	IsLocked            bool       `json:"is_locked"`
	LockedUntil         *time.Time `json:"locked_until,omitempty"`
	FailedLoginAttempts int        `json:"failed_login_attempts"`
	CreatedAt           time.Time  `json:"created_at"`
	LastLogin           *time.Time `json:"last_login,omitempty"`
	ActiveSessions      int        `json:"active_sessions"`
}