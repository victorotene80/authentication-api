package commands

type ChangePasswordCommand struct {
    UserID      string
    OldPassword string
    NewPassword string
    IPAddress   string
    UserAgent   string
}

func(c ChangePasswordCommand) CommandName() string {
	return "ChangePasswordCommand"
}