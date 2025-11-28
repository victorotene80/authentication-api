package commands

type LoginUserCommand struct {
    UsernameOrEmail string
    Password        string
    IPAddress       string
    UserAgent       string
}

func(c LoginUserCommand) CommandName() string{
	return "LoginUserCommand"
}