package oauth

type OAuthLoginRequest struct {
	Provider    string `json:"provider" validate:"required,oneof=google github microsoft"`
	RedirectURI string `json:"redirect_uri" validate:"required,url"`
}