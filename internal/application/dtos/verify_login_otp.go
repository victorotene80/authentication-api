package dtos

type VerifyLoginOTPResult struct {
	UserID       string
	Email        string
	Username     string
	FirstName    string
	LastName     string
	Role         string
	AccessToken  string
	RefreshToken string
	ExpiresAt    string
	ExpiresIn    int64
	SessionID    string
}