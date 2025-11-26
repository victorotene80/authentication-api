package response

import (
	"authentication/api/http/dtos/user"
	"time"
)

type OAuthLoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    time.Time    `json:"expires_at"`
	User         user.UserDTO `json:"user"`
}
