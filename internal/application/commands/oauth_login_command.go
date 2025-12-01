package commands

type LoginWithOAuthCommand struct {
    OAuthProvider string
    OAuthID       string
    Email         string 
    IPAddress     string
    UserAgent     string
}

func (c LoginWithOAuthCommand) CommandName() string {
    return "LoginWithOAuthCommand"
}
