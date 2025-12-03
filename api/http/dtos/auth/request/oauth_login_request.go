package request

import (
	"authentication/internal/application/commands"

	"github.com/go-playground/validator/v10"
)

type OAuthLoginRequest struct {
	IDToken       string `json:"id_token" validate:"required"`
	//OAuthProvider string `json:"oauth_provider" validate:"required,oneof=google github facebook apple"`
	AccessToken   string `json:"access_token,omitempty"` // Optional, some providers need it
}

func (r *OAuthLoginRequest) Validate(v *validator.Validate) error {
	return v.Struct(r)
}

func (r *OAuthLoginRequest) ToCommand(ip, ua string) commands.LoginOAuthUserCommand {
	return commands.LoginOAuthUserCommand{
		OAuthProvider: "google",
		IDToken:       r.IDToken,
		AccessToken:   r.AccessToken,
		IPAddress:     ip,
		UserAgent:     ua,
	}
}
