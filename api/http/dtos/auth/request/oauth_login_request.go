package request

import "authentication/internal/application/commands"

type OAuthLoginRequest struct {
    OAuthProvider string `json:"oauth_provider" validate:"required,oneof=google github facebook apple"`
    OAuthID       string `json:"oauth_id" validate:"required"`
    Email         string `json:"email" validate:"required,email"`
}

func(r *OAuthLoginRequest) ToCommand(ipAddress, userAgent string) commands.LoginWithOAuthCommand {
	return commands.LoginWithOAuthCommand{
		OAuthProvider: r.OAuthProvider,
		OAuthID:       r.OAuthID,
		Email:         r.Email,
	}
}
