package commands

type LoginWithEmailCommand struct {
    Email     string
    Password  string
    IPAddress string
    UserAgent string
}

func (c LoginWithEmailCommand) CommandName() string {
    return "LoginWithEmailCommand"
}
