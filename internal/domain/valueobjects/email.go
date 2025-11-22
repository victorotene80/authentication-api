package valueobjects

import (
	"regexp"
	"strings"
	"authentication/internal/domain"
)

type Email struct {
	value string
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func NewEmail(email string) (Email, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	if email == "" {
		return Email{}, domain.ErrEmptyEmail
	}

	if !emailRegex.MatchString(email) {
		return Email{}, domain.ErrInvalidEmailFormat
	}

	return Email{value: email}, nil
}

func (e Email) IsValid() bool {
	return emailRegex.MatchString(e.value)
}

func (e Email) String() string {
	return e.value
}

func (e Email) Equals(other Email) bool {
	return e.value == other.value
}
