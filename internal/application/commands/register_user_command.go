package commands

import (
	"authentication/internal/domain"
)

type RegisterUserCommand struct {
	Email     string
	Username  string
	Password  string
	FirstName string
	LastName  string
	Phone     string // optional
}

func (c RegisterUserCommand) CommandName() string {
	return "RegisterUserCommand"
}

// NOTE: We don’t do full validation here anymore.
// Use Value Objects + AuthValidator in the handler.
func (c RegisterUserCommand) Validate() error {
	if c.Email == "" {
		return domain.ErrEmptyEmail
	}
	if c.Username == "" {
		return domain.ErrEmptyUsername
	}
	if c.Password == "" {
		return domain.ErrEmptyPassword
	}
	if c.FirstName == "" {
		return domain.ErrFirstNameRequired
	}
	if c.LastName == "" {
		return domain.ErrLastNameRequired
	}
	return nil
}
