package oauth

import "time"

type OAuthClientDTO struct {
	ID        string    `json:"id"`
	ClientID  string    `json:"client_id"`
	Provider  string    `json:"provider"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}
