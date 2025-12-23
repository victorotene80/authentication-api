package commands

type VerifyLoginOTPCommand struct {
	Email     string
	OTPCode   string
	IPAddress string
	UserAgent string
	DeviceID  string
}