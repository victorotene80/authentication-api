package requests

type CancelDeletionRequest struct {
	Token string `json:"token" validate:"required"`
}