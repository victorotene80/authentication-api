package requests

type RevokeSessionRequest struct {
	SessionID string `json:"session_id" validate:"required,uuid"`
}