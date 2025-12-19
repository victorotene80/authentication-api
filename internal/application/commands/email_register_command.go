package commands

type RegisterEmailUserCommand struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	Username  string
	Role      string
	IPAddress string
	UserAgent string
	DeviceID  string
}

func (c RegisterEmailUserCommand) CommandName() string {
	return "RegisterEmailUser"
}
