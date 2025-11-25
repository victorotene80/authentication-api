package requests

type VerifyPhoneRequest struct {
	Code string `json:"code" validate:"required,len=6,numeric"`
}
