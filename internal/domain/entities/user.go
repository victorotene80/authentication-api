package entities

import (
	"time"

	"authentication/internal/domain/valueobjects"

	"github.com/google/uuid"
)

type UserStatus string

const (
	UserStatusPending  UserStatus = "pending"
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusLocked   UserStatus = "locked"
)

type User struct {
	ID           string
	Username     valueobjects.Username
	Email        valueobjects.Email
	PasswordHash string
	Phone        valueobjects.PhoneNumber
	FirstName    string
	LastName     string
	Role         valueobjects.Role
	IsActive     bool
	IsVerified   bool
	LastLoginAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(
	username valueobjects.Username,
	email valueobjects.Email,
	passwordHash string,
	phone valueobjects.PhoneNumber,
	firstName, lastName string,
	role valueobjects.Role,
) *User {
	now := time.Now()
	return &User{
		ID:           uuid.New().String(),
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Phone:        phone,
		FirstName:    firstName,
		LastName:     lastName,
		Role:         role,
		IsActive:     true,
		IsVerified:   false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (u *User) Deactivate() {
	u.IsActive = false
	u.UpdatedAt = time.Now()
}

func (u *User) Activate() {
	u.IsActive = true
	u.UpdatedAt = time.Now()
}

func (u *User) VerifyEmail() {
	u.IsVerified = true
	u.UpdatedAt = time.Now()
}

func (u *User) UpdatePassword(passwordHash string) {
	u.PasswordHash = passwordHash
	u.UpdatedAt = time.Now()
}

func (u *User) UpdateProfile(firstName, lastName string, phone valueobjects.PhoneNumber) {
	u.FirstName = firstName
	u.LastName = lastName
	u.Phone = phone
	u.UpdatedAt = time.Now()
}

func (u *User) RecordLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.UpdatedAt = now
}

/*
func (u *User) VerifyEmail() error {
	if u.emailVerified {
		return domain.ErrEmailAlreadyInUse
	}

	now := utils.NowUTC()
	u.emailVerified = true
	u.emailVerifiedAt = &now
	u.status = UserStatusActive
	u.updatedAt = now

	// Fire EmailVerifiedEvent (example, you can implement similarly to UserCreatedEvent)
	// u.AddEvent(events.NewEmailVerifiedEvent(u.id, u.email.String()))

	return nil
}

func (u *User) RecordSuccessfulLogin() {
	now := utils.NowUTC()
	u.lastLogin = &now
	u.failedLoginAttempts = 0
	u.lastFailedLogin = nil
	u.updatedAt = now
}


func (u *User) Activate() error {
	if u.status == UserStatusActive {
		return domain.ErrUserAlreadyActive
	}
	u.status = UserStatusActive
	u.updatedAt = utils.NowUTC()
	return nil
}

func (u *User) Deactivate() error {
	if u.status == UserStatusInactive {
		return domain.ErrUserAlreadyInactive
	}
	u.status = UserStatusInactive
	u.updatedAt = utils.NowUTC()
	return nil
}

func (u *User) VerifyPassword(password string) bool {
	return u.password.Verify(password)
}

func (u *User) RecordFailedLogin() error {
	now := utils.NowUTC()
	u.failedLoginAttempts++
	u.lastFailedLogin = &now
	u.updatedAt = now

	if u.failedLoginAttempts >= 5 {
		lockUntil := now.Add(30 * time.Minute)
		u.isLocked = true
		u.lockedUntil = &lockUntil
		u.status = UserStatusLocked
		return domain.ErrTooManyFailedLogins
	}

	return nil
}

func (u *User) UnlockAccount() error {
	if !u.isLocked {
		return nil
	}

	u.isLocked = false
	u.lockedUntil = nil
	u.failedLoginAttempts = 0
	u.lastFailedLogin = nil
	u.status = UserStatusActive
	u.updatedAt = utils.NowUTC()
	return nil
}

*/
