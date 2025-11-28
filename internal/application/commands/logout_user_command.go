package commands

type LogoutUserCommand struct {
    UserID       string
    SessionID    string
    RefreshToken string
    IPAddress    string
    UserAgent    string
}

func (c LogoutUserCommand) CommandName() string {
	return "LogoutUserCommand"
}