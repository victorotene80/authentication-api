package oauth

type LinkOAuthAccountRequest struct {
	Provider string `json:"provider" validate:"required,oneof=google github microsoft"`
	Code     string `json:"code" validate:"required"`
	State    string `json:"state" validate:"required"`
}