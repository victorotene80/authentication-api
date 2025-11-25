package requests

type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}