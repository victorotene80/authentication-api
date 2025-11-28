package commands

type UpdateUserCommand struct {
    UserID    string
    FirstName string
    LastName  string
    Phone     string
    IPAddress string
    UserAgent string
}

func (c UpdateUserCommand) CommandName() string {
	return "UpdateUserCommand"
}