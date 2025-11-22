package repositories

import "errors"

var (
	ErrNotFound       = errors.New("record not found")
	ErrAlreadyExists  = errors.New("record already exists")
	ErrConflict       = errors.New("conflicting record state")
	ErrInvalidInput   = errors.New("invalid input for repository operation")
	ErrDBUnavailable  = errors.New("database unavailable")
	ErrTxFailed       = errors.New("transaction failed")
	ErrTimeout        = errors.New("repository operation timed out")
	ErrConstraintViolation = errors.New("data constraint violation")
	ErrSerialization       = errors.New("serialization or marshaling error")
	ErrPermissionDenied = errors.New("permission denied for repository access")
)
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrNotFound)
}
