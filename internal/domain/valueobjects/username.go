package valueobjects

import (
	"authentication/internal/domain"
	"regexp"
	"strings"
)

type Username struct {
	value string
}

func NewUsername(username string) (Username, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return Username{}, domain.ErrEmptyUsername
	}

	if len(username) < 3 {
		return Username{}, domain.ErrUsernameTooShort
	}

	if len(username) > 30 {
		return Username{}, domain.ErrUsernameTooLong
	}

	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !usernameRegex.MatchString(username) {
		return Username{}, domain.ErrInvalidUsernameFormat
	}

	return Username{value: username}, nil
}

func (u Username) String() string {
	return u.value
}

func (u Username) IsEmpty() bool {
	return u.value == ""
}
