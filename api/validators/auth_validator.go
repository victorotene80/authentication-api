// internal/application/validators/auth_validator.go
package validators

import (
    "authentication/internal/application/commands"
    "authentication/internal/domain/valueobjects"
    "errors"
    "strings"
)

type AuthValidator struct {
    commonPasswords map[string]bool
}

func NewAuthValidator() *AuthValidator {
    return &AuthValidator{
        commonPasswords: map[string]bool{
            "password123": true,
            "password":    true,
            "12345678":    true,
            "qwerty123":   true,
        },
    }
}

func (v *AuthValidator) ValidateRegistration(cmd commands.RegisterUserCommand) error {

    // Email (via VO)
    if _, err := valueobjects.NewEmail(cmd.Email); err != nil {
        return err
    }

    // Username (via VO)
    if _, err := valueobjects.NewUsername(cmd.Username); err != nil {
        return err
    }

    // Password (via VO)
    if _, err := valueobjects.NewPassword(cmd.Password); err != nil {
        return err
    }

    // Names
    if err := v.validateName(cmd.FirstName, "first name"); err != nil {
        return err
    }
    if err := v.validateName(cmd.LastName, "last name"); err != nil {
        return err
    }

    // Optional phone
    if strings.TrimSpace(cmd.Phone) != "" {
        if _, err := valueobjects.NewPhoneNumber(cmd.Phone); err != nil {
            return err
        }
    }

    // Additional business-specific checks
    if v.commonPasswords[strings.ToLower(cmd.Password)] {
        return errors.New("password too common")
    }

    return nil
}
