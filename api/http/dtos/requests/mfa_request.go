package requests

type EnableMFARequest struct {
	Password string `json:"password" validate:"required"`
}

type ConfirmMFARequest struct {
	Code string `json:"code" validate:"required,len=6,numeric"`
}

type DisableMFARequest struct {
	Password string `json:"password" validate:"required"`
	Code     string `json:"code" validate:"required,len=6,numeric"`
}

type VerifyMFARequest struct {
	Code string `json:"code" validate:"required,len=6,numeric"`
}
