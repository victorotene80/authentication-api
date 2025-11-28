package request

import (
	"fmt"

	"github.com/go-playground/validator/v10"

	"authentication/internal/application/commands"
)

type EmailRegistrationRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=12,max=128"`
	Phone     string `json:"phone" validate:"required,e164"`
	FirstName string `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string `json:"last_name" validate:"required,min=1,max=100"`
	Username  string `json:"username" validate:"required,min=3,max=50,alphanum"`
}

func (r *EmailRegistrationRequest) Validate(v *validator.Validate) error {
	if v == nil {
		return fmt.Errorf("validator is nil")
	}
	return v.Struct(r)
}

func (r *EmailRegistrationRequest) ToCommand(ipAddress, userAgent string) commands.RegisterUserCommand {
	return commands.RegisterUserCommand{
		Username:      r.Username,
		Email:         r.Email,
		Password:      r.Password,
		Phone:         r.Phone,
		FirstName:     r.FirstName,
		LastName:      r.LastName,
		Role:          "user",
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		OAuthProvider: "",
		OAuthID:       "",
		IsOAuth:       false,
	}
}
