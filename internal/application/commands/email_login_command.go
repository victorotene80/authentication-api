package commands

type LoginEmailUserCommand struct {
    Email     string
    Password  string
    IPAddress string
    UserAgent string
}

func (c LoginEmailUserCommand) CommandName() string {
    return "LoginWithEmailCommand"
}
