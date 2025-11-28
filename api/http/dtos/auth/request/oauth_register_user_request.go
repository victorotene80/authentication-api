package request

import (
	"fmt"

	"github.com/go-playground/validator/v10"

	"authentication/internal/application/commands"
)

type OAuthRegistrationRequest struct {
	OAuthProvider string `json:"oauth_provider" validate:"required,oneof=google github facebook apple"`
	OAuthID       string `json:"oauth_id" validate:"required"`
	Email         string `json:"email" validate:"required,email"`
}

func (r *OAuthRegistrationRequest) Validate(v *validator.Validate) error {
	if v == nil {
		return fmt.Errorf("validator is nil")
	}
	return v.Struct(r)
}

func (r *OAuthRegistrationRequest) ToCommand(ipAddress, userAgent string) commands.RegisterUserCommand {
	return commands.RegisterUserCommand{
		Username:      "",  
		Email:         r.Email,
		Password:      "", 
		Phone:         "",
		FirstName:     "",
		LastName:      "",
		Role:          "user",
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		OAuthProvider: r.OAuthProvider,
		OAuthID:       r.OAuthID,
		IsOAuth:       true,
	}
}
