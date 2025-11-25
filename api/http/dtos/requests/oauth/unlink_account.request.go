package oauth

type UnlinkOAuthAccountRequest struct {
	Provider string `json:"provider" validate:"required,oneof=google github microsoft"`
	Password string `json:"password" validate:"required"`
}