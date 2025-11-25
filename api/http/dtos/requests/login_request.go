package requests

type LoginRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required,min=8"`
	DeviceInfo *DeviceInfoRequest `json:"device_info,omitempty"`
	MFACode    string `json:"mfa_code,omitempty" validate:"omitempty,len=6,numeric"`
}