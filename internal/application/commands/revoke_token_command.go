package commands

type RevokeTokenCommand struct {
    UserID    string
    Token     string
    IPAddress string
    UserAgent string
}

func (c RevokeTokenCommand) CommandName() string {
	return "RevokeTokenCommand"
}