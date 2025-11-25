package requests

type RequestPasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}
