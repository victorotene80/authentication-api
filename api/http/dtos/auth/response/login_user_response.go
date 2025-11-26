package response

import (
	"time"
	"authentication/api/http/dtos/user"
)

type LoginUserResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         user.UserDTO   `json:"user"`
}
