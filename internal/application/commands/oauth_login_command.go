package commands

type OAuthLoginCommand struct {
    Provider    string
    Code        string
    RedirectURI string
    State       string
    IPAddress   string
    UserAgent   string
}

func (c OAuthLoginCommand) CommandName() string {
	return "OAuthLoginCommand"
}