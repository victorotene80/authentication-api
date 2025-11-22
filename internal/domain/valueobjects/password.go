package valueobjects

import (
	"authentication/internal/domain"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/crypto/argon2"
)

type Password struct {
	hashedValue string
}

const (
	Argon2Time    uint32 = 3         // Number of iterations
	Argon2Memory  uint32 = 64 * 1024 // 64 MB
	Argon2Threads uint8  = 4         // Parallelism
	Argon2KeyLen  uint32 = 32        // Key length
	Argon2SaltLen int    = 16        // Salt length
)

func NewPassword(plainPassword string) (Password, error) {
	if plainPassword == "" {
		return Password{}, domain.ErrEmptyPassword
	}
	
	if err := ValidatePasswordStrength(plainPassword); err != nil {
		return Password{}, err
	}

	salt, err := generateSalt(Argon2SaltLen)
	if err != nil {
		return Password{}, fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(plainPassword), salt, Argon2Time, Argon2Memory, Argon2Threads, Argon2KeyLen)

	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		Argon2Memory,
		Argon2Time,
		Argon2Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return Password{hashedValue: encodedHash}, nil
}

func NewPasswordFromHash(hashedValue string) (Password, error) {
	if hashedValue == "" {
		return Password{}, domain.ErrEmptyPassword
	}
	if !strings.HasPrefix(hashedValue, "$argon2id$") {
		return Password{}, domain.ErrPasswordTooWeak
	}
	return Password{hashedValue: hashedValue}, nil
}

func (p Password) IsEmpty() bool {
	return p.hashedValue == ""
}

func (p Password) Verify(plainPassword string) bool {
	parts := strings.Split(p.hashedValue, "$")
	if len(parts) != 6 {
		return false
	}

	var version int
	var memory, time uint32
	var threads uint8

	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false
	}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}
	storedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}

	computedHash := argon2.IDKey([]byte(plainPassword), salt, time, memory, threads, uint32(len(storedHash)))
	return subtle.ConstantTimeCompare(storedHash, computedHash) == 1
}

func (p Password) Hash() string {
	return p.hashedValue
}

func (p Password) NeedsRehash() bool {
	parts := strings.Split(p.hashedValue, "$")
	if len(parts) != 6 {
		return true
	}

	var memory, time uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return true
	}

	return memory < Argon2Memory || time < Argon2Time || threads < Argon2Threads
}

type PasswordStrengthRequirements struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireDigit   bool
	RequireSpecial bool
	MaxLength      int
}

var DefaultPasswordRequirements = PasswordStrengthRequirements{
	MinLength:      12,
	RequireUpper:   true,
	RequireLower:   true,
	RequireDigit:   true,
	RequireSpecial: true,
	MaxLength:      128,
}

func ValidatePasswordStrength(password string) error {
	return ValidatePasswordWithRequirements(password, DefaultPasswordRequirements)
}

func ValidatePasswordWithRequirements(password string, reqs PasswordStrengthRequirements) error {
	if len(password) < reqs.MinLength {
		return domain.ErrPasswordTooShort
	}
	if len(password) > reqs.MaxLength {
		return domain.ErrPasswordTooLong
	}

	if reqs.RequireUpper && !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return domain.ErrPasswordTooWeak
	}
	if reqs.RequireLower && !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return domain.ErrPasswordTooWeak
	}
	if reqs.RequireDigit && !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return domain.ErrPasswordTooWeak
	}
	if reqs.RequireSpecial && !regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>_\-+=\[\]\\\/;']`).MatchString(password) {
		return domain.ErrPasswordTooWeak
	}

	return nil
}

func generateSalt(length int) ([]byte, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}
