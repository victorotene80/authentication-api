package request

import (
	"authentication/internal/application/commands"

	"github.com/go-playground/validator/v10"
)

type OAuthRegistrationRequest struct {
	IDToken       string `json:"id_token" validate:"required"`
	OAuthProvider string `json:"oauth_provider" validate:"required,oneof=google github facebook apple"`

	//OAuthProvider string `json:"oauth_provider" validate:"required,oneof=google github facebook apple"`
	//OAuthID       string `json:"oauth_id" validate:"required"`
	//Email         string `json:"email" validate:"required,email"`
}

func (r *OAuthRegistrationRequest) Validate(v *validator.Validate) error {
	return v.Struct(r)
}

func (r *OAuthRegistrationRequest) ToCommand(ip, ua string) commands.RegisterOAuthUserCommand {
	return commands.RegisterOAuthUserCommand{
		OAuthProvider: "google",
		IDToken:       r.IDToken,
		AccessToken:   "", 
		IPAddress:     ip,
		UserAgent:     ua,
		//IsOAuth:       true,
		Role:          "user", 
	}
}
