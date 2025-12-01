package response

import "time"

type RegisterUserResponse struct {
	UserID       string     `json:"user_id"`
	//Username     string     `json:"username"`
	Email        string     `json:"email"`
	FirstName    string     `json:"first_name"`
	LastName     string     `json:"last_name"`
	Role         string     `json:"role"`
	CreatedAt    time.Time  `json:"created_at"`
	AccessToken  string     `json:"access_token,omitempty"`
	RefreshToken string     `json:"refresh_token,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}
