package crypto

import "errors"

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrTokenNotYetValid = errors.New("token not yet valid")
)
