package dtos

type RegisterOAuthUserResult struct {
	UserID          string
	Email           string
	Username        string 
	FirstName       string
	LastName        string
	Role            string
	IsOAuthUser     bool   
	OAuthProvider   string 
	RequiresOnboard bool   
	IsNewUser       bool   
	AccessToken     string
	RefreshToken    string
	ExpiresAt       string
	ExpiresIn       int64
	SessionID       string
}
