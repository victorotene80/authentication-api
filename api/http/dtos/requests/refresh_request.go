package requests

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
	DeviceInfo   *DeviceInfoRequest `json:"device_info,omitempty"`
}