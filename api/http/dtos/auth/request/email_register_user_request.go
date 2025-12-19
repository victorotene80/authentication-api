package request

import (
	"authentication/internal/application/commands"

	"github.com/go-playground/validator/v10"
)

type EmailRegistrationRequest struct {
	FirstName string `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string `json:"last_name" validate:"required,min=2,max=100"`
	Username  string `json:"username" validate:"omitempty,min=3,max=50,alphanum"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8,max=200"`
}

func (r *EmailRegistrationRequest) Validate(v *validator.Validate) error {
	return v.Struct(r)
}

func (r *EmailRegistrationRequest) ToCommand(ip, ua string) commands.RegisterEmailUserCommand {
	return commands.RegisterEmailUserCommand{
		Email:     r.Email,
		Password:  r.Password,
		FirstName: r.FirstName,
		LastName:  r.LastName,
		Username:  r.Username,
		IPAddress: ip,
		UserAgent: ua,
		//IsOAuth:   false,
		Role:      "user",
	}
}
