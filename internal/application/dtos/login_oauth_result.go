package dtos

type LoginOAuthUserResult struct {
	UserID           string
	Email            string
	FirstName        string
	LastName         string
	Role             string
	RequiresOTP      bool // True if user is email-only and needs OTP
	IsNewUser        bool
	OAuthProvider    string
	RequiresOnboard  bool
	AccessToken      string
	RefreshToken     string
	ExpiresAt        string
	ExpiresIn        int64
	SessionID        string
	OTPSent          bool
	Message          string
}
