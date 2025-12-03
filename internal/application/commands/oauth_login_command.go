package commands

type LoginOAuthUserCommand struct {
	OAuthProvider string
	IDToken       string
	AccessToken   string
	IPAddress     string
	UserAgent     string
}

func (c LoginOAuthUserCommand) CommandName() string {
	return "LoginOAuthUser"
}
