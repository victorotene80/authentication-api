package requests

type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
	Type  string `json:"type" validate:"required,oneof=email phone"`
}