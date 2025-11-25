package requests

type DeviceInfoRequest struct {
	Fingerprint string `json:"fingerprint,omitempty"`
	UserAgent   string `json:"user_agent,omitempty"`
	Platform    string `json:"platform,omitempty"`
	Browser     string `json:"browser,omitempty"`
	DeviceType  string `json:"device_type,omitempty" validate:"omitempty,oneof=mobile tablet desktop"`
	TrustDevice bool   `json:"trust_device,omitempty"`
}