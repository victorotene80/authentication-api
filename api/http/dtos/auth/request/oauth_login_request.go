package request

import (
	"authentication/internal/application/commands"

	"github.com/go-playground/validator/v10"
)

type LoginOAuthUserCommand struct {
	OAuthProvider string
	IDToken       string
	AccessToken   string
	Email         string
	IPAddress     string
	UserAgent     string
	DeviceID      string
}


func (r *LoginOAuthUserCommand) Validate(v *validator.Validate) error {
	return v.Struct(r)
}

func (r *LoginOAuthUserCommand) ToCommand(ip, ua string) commands.LoginOAuthUserCommand {
	return commands.LoginOAuthUserCommand{
		OAuthProvider: "google",
		IDToken:       r.IDToken,
		AccessToken:   r.AccessToken,
		IPAddress:     ip,
		UserAgent:     ua,
	}
}
