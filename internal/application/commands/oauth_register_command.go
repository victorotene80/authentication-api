package commands

type RegisterOAuthUserCommand struct {
	OAuthProvider string
	IDToken       string
	AccessToken   string
	Role          string
	IPAddress     string
	UserAgent     string
}

func (c RegisterOAuthUserCommand) CommandName() string {
	return "RegisterOAuthUser"
}
