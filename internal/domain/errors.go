package domain

import "errors"

var (
	ErrInvalidEmailFormat            = errors.New("invalid email format")
	ErrEmptyEmail                    = errors.New("email is required")
	ErrEmailAlreadyInUse             = errors.New("email is already in use")
	ErrEmptyPassword                 = errors.New("password is required")
	ErrPasswordTooWeak               = errors.New("password does not meet policy")
	ErrPasswordTooShort              = errors.New("password is too short")
	ErrPasswordTooLong               = errors.New("password is too long")
	ErrUserLocked                    = errors.New("user account is locked")
	ErrInvalidID                     = errors.New("invalid id")
	ErrInvalidPhoneFormat            = errors.New("invalid phone number format")
	ErrEmptyPhoneNumber              = errors.New("phone number is required")
	ErrInvalidUsernameFormat         = errors.New("username can only contain letters, numbers, underscores, and hyphens")
	ErrEmptyUsername                 = errors.New("username is required")
	ErrUsernameTooShort              = errors.New("username must be at least 3 characters")
	ErrUsernameTooLong               = errors.New("username must not exceed 30 characters")
	ErrTooManyFailedLogins           = errors.New("too many failed login attempts, account is locked")
	ErrInactiveUser                  = errors.New("user account is inactive")
	ErrUserAlreadyActive             = errors.New("user account is already active")
	ErrUserAlreadyInactive           = errors.New("user account is already inactive")
	ErrUsernameAlreadyTaken          = errors.New("username already taken")
	ErrFirstNameRequired             = errors.New("first name is required")
	ErrLastNameRequired              = errors.New("last name is required")
	ErrInvalidSession                = errors.New("invalid session")
	ErrSessionExpired                = errors.New("session expired")
	ErrCannonotLogoutInactiveSession = errors.New("cannot logout an inactive session")
	ErrCannonotRevokeInactiveSession = errors.New("cannot revoke an inactive session")

	ErrUserNotFound         = errors.New("user not found")
	ErrUsernameAlreadyInUse = errors.New("username already in use")
	ErrInvalidEmail         = errors.New("invalid email format")
	ErrInvalidUsername      = errors.New("invalid username format")
	ErrInvalidPassword      = errors.New("invalid password format")
	ErrInvalidRole          = errors.New("invalid role")
	ErrInvalidPhone         = errors.New("invalid phone number")
	// Authentication errors
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrPasswordMismatch   = errors.New("password does not match")
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionRevoked     = errors.New("session revoked")

	// OAuth errors
	ErrOAuthVerificationFailed   = errors.New("oauth verification failed")
	ErrOAuthEmailNotVerified     = errors.New("oauth email not verified")
	ErrOAuthProviderNotSupported = errors.New("oauth provider not supported")
	
	// Authentication errors
	ErrInvalidToken         = errors.New("invalid token")
	ErrTokenExpired         = errors.New("token expired")
	
	// OAuth errors
	ErrOAuthProviderMismatch     = errors.New("email registered with different oauth provider")

)
