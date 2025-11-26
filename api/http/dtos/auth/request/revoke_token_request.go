package request

type RevokeTokenRequest struct {
	Token string `json:"token" validate:"required"`
}
