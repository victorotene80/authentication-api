package admin

import "time"

type AdminUserStatistics struct {
	TotalLogins             int        `json:"total_logins"`
	FailedLogins            int        `json:"failed_logins"`
	SuccessfulLogins        int        `json:"successful_logins"`
	SuspiciousLoginAttempts int        `json:"suspicious_login_attempts"`
	PasswordChanges         int        `json:"password_changes"`
	LastPasswordChange      *time.Time `json:"last_password_change,omitempty"`
}
