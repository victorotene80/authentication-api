package services

import (
	"authentication/internal/domain"
	"regexp"
)

type PasswordStrengthRequirements struct {
	MinLength      int
	MaxLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireDigit   bool
	RequireSpecial bool
}

var DefaultPasswordRequirements = PasswordStrengthRequirements{
	MinLength:      12,
	MaxLength:      128,
	RequireUpper:   true,
	RequireLower:   true,
	RequireDigit:   true,
	RequireSpecial: true,
}

type PasswordPolicyService struct {
	requirements PasswordStrengthRequirements
}

func NewPasswordPolicyService() PasswordPolicyService {
	return PasswordPolicyService{requirements: DefaultPasswordRequirements}
}

func (ps PasswordPolicyService) Validate(password string) error {
	r := ps.requirements

	if len(password) < r.MinLength {
		return domain.ErrPasswordTooShort
	}
	if len(password) > r.MaxLength {
		return domain.ErrPasswordTooLong
	}

	if r.RequireUpper && !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return domain.ErrPasswordTooWeak
	}
	if r.RequireLower && !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return domain.ErrPasswordTooWeak
	}
	if r.RequireDigit && !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return domain.ErrPasswordTooWeak
	}
	if r.RequireSpecial && !regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>_\-+=\[\]\\\/;']`).MatchString(password) {
		return domain.ErrPasswordTooWeak
	}

	return nil
}
