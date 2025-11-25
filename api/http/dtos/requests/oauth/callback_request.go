package oauth

type OAuthCallbackRequest struct {
	Code  string `json:"code" validate:"required"`
	State string `json:"state" validate:"required"`
}