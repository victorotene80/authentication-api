package commands

type RefreshTokenCommand struct {
    RefreshToken string
    IPAddress    string
    UserAgent    string
}
func (c RefreshTokenCommand) CommandName() string {
	return "RefreshTokenCommand"
}