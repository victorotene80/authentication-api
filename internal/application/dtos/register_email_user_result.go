package dtos

type RegisterEmailUserResult struct {
	UserID          string
	Email           string
	Username        string
	FirstName       string
	LastName        string
	Role            string
	IsOAuthUser     bool   // Always false for email registration
	OAuthProvider   string // Always empty for email registration
	RequiresOnboard bool   // Always false - email users complete registration upfront
	IsNewUser       bool   // Always true - this is registration
}