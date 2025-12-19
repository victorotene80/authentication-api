package commands

type RegisterOAuthUserCommand struct {
	OAuthProvider string
	Email         string
	IDToken       string
	AccessToken   string
	Role          string
	IPAddress     string
	UserAgent     string
	DeviceID      string
}
