package valueobjects

import (
	"regexp"
	"strings"
	"authentication/internal/domain"
)

type PhoneNumber struct {
	value string
}

var phoneRegex = regexp.MustCompile(`^\+[1-9]\d{6,14}$`)

func NewPhoneNumber(phone string) (PhoneNumber, error) {
	phone = strings.TrimSpace(phone)

	/*if phone == "" {
		return PhoneNumber{}, domain.ErrEmptyPhoneNumber
	}*/

	if !strings.HasPrefix(phone, "+") {
		phone = "+" + phone
	}

	if !phoneRegex.MatchString(phone) {
		return PhoneNumber{}, domain.ErrInvalidPhoneFormat
	}

	return PhoneNumber{value: phone}, nil
}

func (p PhoneNumber) String() string {
	return p.value
}

func (p PhoneNumber) Equals(other PhoneNumber) bool {
	return p.value == other.value
}
