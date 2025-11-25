package responses

import "time"

type UserProfileResponse struct {
	User             UserResponse            `json:"user"`
	Statistics       UserStatistics          `json:"statistics"`
	SecuritySettings SecuritySettings        `json:"security_settings"`
	LinkedAccounts   []LinkedAccountResponse `json:"linked_accounts,omitempty"`
}

type UserStatistics struct {
	TotalLogins         int        `json:"total_logins"`
	FailedLoginAttempts int        `json:"failed_login_attempts"`
	ActiveSessions      int        `json:"active_sessions"`
	LastPasswordChange  *time.Time `json:"last_password_change,omitempty"`
}

type SecuritySettings struct {
	TwoFactorEnabled    bool   `json:"two_factor_enabled"`
	TrustedDevicesCount int    `json:"trusted_devices_count"`
	PasswordStrength    string `json:"password_strength"`
	AccountLocked       bool   `json:"account_locked"`
}