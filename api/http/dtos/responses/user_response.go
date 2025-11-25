package responses

import "time"

type UserResponse struct {
	ID               string     `json:"id"`
	Email            string     `json:"email"`
	Username         string     `json:"username"`
	FirstName        string     `json:"first_name"`
	LastName         string     `json:"last_name"`
	FullName         string     `json:"full_name"`
	Phone            string     `json:"phone,omitempty"`
	EmailVerified    bool       `json:"email_verified"`
	PhoneVerified    bool       `json:"phone_verified,omitempty"`
	Status           string     `json:"status"`
	Role             string     `json:"role"`
	TwoFactorEnabled bool       `json:"two_factor_enabled"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	LastLogin        *time.Time `json:"last_login,omitempty"`
}