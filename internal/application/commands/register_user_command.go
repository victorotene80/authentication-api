package commands

type RegisterUserCommand struct {
	Email         string
	Password      string
	FirstName     string
	LastName      string
	Username      string
	IsOAuth       bool
	OAuthProvider string
	IDToken       string
	AccessToken   string
	IPAddress     string
	UserAgent     string
	Role          string
}

func (c RegisterUserCommand) IsOAuthRegistration() bool {
	return c.IsOAuth
}

func (c RegisterUserCommand) CommandName() string {
	return "RegisterUserCommand"
}
