package dtos

type LoginEmailUserResult struct {
	UserID           string
	Email            string
	Username         string
	FirstName        string
	LastName         string
	Role             string
	RequiresOTP      bool // True if user is OAuth-only and needs OTP
	AccessToken      string
	RefreshToken     string
	ExpiresAt        string
	ExpiresIn        int64
	SessionID        string
	OTPSent          bool // True if OTP was sent
	Message          string
}