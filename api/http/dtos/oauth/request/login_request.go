package request

type OAuthLoginRequest struct {
    Provider     string `json:"provider" validate:"required,oneof=google github"`
    Code         string `json:"code" validate:"required"`
    RedirectURI  string `json:"redirect_uri" validate:"required,url"`
    State        string `json:"state,omitempty"`
}