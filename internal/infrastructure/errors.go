package infrastructure

import "errors"

var (
	ErrHandlerNotFound   = errors.New("handler not found")
	ErrInvalidCommandName = errors.New("invalid command name")
	ErrAlreadyRegistered = errors.New("handler already registered")
	ErrBusSealed         = errors.New("command bus sealed, no further registrations allowed")
)
