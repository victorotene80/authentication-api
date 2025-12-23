package commands

type LoginOAuthUserCommand struct {
	OAuthProvider string
	IDToken       string
	AccessToken   string
	Email         string
	IPAddress     string
	UserAgent     string
	DeviceID      string
}
