package services

import (
	"authentication/internal/domain"
	"authentication/internal/domain/valueobjects"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	Argon2Time    uint32 = 3
	Argon2Memory  uint32 = 64 * 1024
	Argon2Threads uint8  = 4
	Argon2KeyLen  uint32 = 32
	Argon2SaltLen int    = 16
)

type PasswordHashingService struct {
	policy PasswordPolicyService
}

func NewPasswordHashingService(policy PasswordPolicyService) PasswordHashingService {
	return PasswordHashingService{policy: policy}
}

func (s PasswordHashingService) HashPassword(plain string) (valueobjects.Password, error) {
	if plain == "" {
		return valueobjects.Password{}, domain.ErrEmptyPassword
	}

	if err := s.policy.Validate(plain); err != nil {
		return valueobjects.Password{}, err
	}

	salt, err := generateSalt(Argon2SaltLen)
	if err != nil {
		return valueobjects.Password{}, fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(plain), salt, Argon2Time, Argon2Memory, Argon2Threads, Argon2KeyLen)

	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		Argon2Memory,
		Argon2Time,
		Argon2Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return valueobjects.NewPassword(encodedHash), nil
}

func (s PasswordHashingService) Verify(plain string, password valueobjects.Password) bool {
	parts := strings.Split(password.Value(), "$")
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

	computed := argon2.IDKey([]byte(plain), salt, time, memory, threads, uint32(len(storedHash)))

	return subtle.ConstantTimeCompare(storedHash, computed) == 1
}

func (s PasswordHashingService) NeedsRehash(password valueobjects.Password) bool {
	parts := strings.Split(password.Value(), "$")
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

func generateSalt(length int) ([]byte, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}
