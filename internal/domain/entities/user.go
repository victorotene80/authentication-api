package entities

import (
	"time"

	"authentication/internal/domain"
	"authentication/internal/domain/events"
	"authentication/internal/domain/valueobjects"
	"authentication/shared/utils"

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
	domain.AggregateRoot // embed AggregateRoot for event handling

	id                  uuid.UUID
	email               valueobjects.Email
	username            valueobjects.Username
	password            valueobjects.Password
	firstName           string
	lastName            string
	phone               valueobjects.PhoneNumber
	emailVerified       bool
	emailVerifiedAt     *time.Time
	status              UserStatus
	isLocked            bool
	lockedUntil         *time.Time
	failedLoginAttempts int
	lastFailedLogin     *time.Time
	twoFactorEnabled    bool
	twoFactorSecret     string
	createdAt           time.Time
	updatedAt           time.Time
	lastLogin           *time.Time
}

// ---------------- Factory / Constructor ----------------

func NewUser(email valueobjects.Email, username valueobjects.Username, password valueobjects.Password, firstName, lastName string) (*User, error) {
	if firstName == "" {
		return nil, domain.ErrFirstNameRequired
	}
	if lastName == "" {
		return nil, domain.ErrLastNameRequired
	}

	now := utils.NowUTC()
	user := &User{
		id:                  uuid.New(),
		email:               email,
		username:            username,
		password:            password,
		firstName:           firstName,
		lastName:            lastName,
		emailVerified:       false,
		status:              UserStatusPending,
		isLocked:            false,
		failedLoginAttempts: 0,
		twoFactorEnabled:    false,
		createdAt:           now,
		updatedAt:           now,
	}

	// Fire domain event via AggregateRoot
	user.AddEvent(events.NewUserCreatedEvent(
		user.id,
		user.email.String(),
		user.username.String(),
	))

	return user, nil
}

// ---------------- Reconstruction from persistence ----------------

func ReconstructUser(
	id uuid.UUID,
	email valueobjects.Email,
	username valueobjects.Username,
	password valueobjects.Password,
	firstName, lastName string,
	phone valueobjects.PhoneNumber,
	emailVerified bool,
	emailVerifiedAt *time.Time,
	status UserStatus,
	isLocked bool,
	lockedUntil *time.Time,
	failedLoginAttempts int,
	lastFailedLogin *time.Time,
	twoFactorEnabled bool,
	twoFactorSecret *string,
	createdAt, updatedAt time.Time,
	lastLogin *time.Time,
) *User {
	secret := ""
	if twoFactorSecret != nil {
		secret = *twoFactorSecret
	}

	return &User{
		id:                  id,
		email:               email,
		username:            username,
		password:            password,
		firstName:           firstName,
		lastName:            lastName,
		phone:               phone,
		emailVerified:       emailVerified,
		emailVerifiedAt:     emailVerifiedAt,
		status:              status,
		isLocked:            isLocked,
		lockedUntil:         lockedUntil,
		failedLoginAttempts: failedLoginAttempts,
		lastFailedLogin:     lastFailedLogin,
		twoFactorEnabled:    twoFactorEnabled,
		twoFactorSecret:     secret,
		createdAt:           createdAt,
		updatedAt:           updatedAt,
		lastLogin:           lastLogin,
	}
}

// ---------------- Accessors ----------------

func (u *User) ID() uuid.UUID                   { return u.id }
func (u *User) Email() valueobjects.Email       { return u.email }
func (u *User) Username() valueobjects.Username { return u.username }
func (u *User) FirstName() string               { return u.firstName }
func (u *User) LastName() string                { return u.lastName }
func (u *User) FullName() string                { return u.firstName + " " + u.lastName }
func (u *User) IsEmailVerified() bool           { return u.emailVerified }
func (u *User) Status() UserStatus              { return u.status }
func (u *User) IsLocked() bool                  { return u.isLocked }
func (u *User) CreatedAt() time.Time            { return u.createdAt }
func (u *User) UpdatedAt() time.Time            { return u.updatedAt }
func (u *User) PasswordHash() string            { return u.password.Hash() }
func (u *User) EmailVerifiedAt() *time.Time     { return u.emailVerifiedAt }
func (u *User) LastFailedLogin() *time.Time     { return u.lastFailedLogin }
func (u *User) Phone() valueobjects.PhoneNumber { return u.phone }
func (u *User) FailedLoginAttempts() int        { return u.failedLoginAttempts }
func (u *User) TwoFactorEnabled() bool          { return u.twoFactorEnabled }
func (u *User) TwoFactorSecret() string         { return u.twoFactorSecret }
func (u *User) LastLogin() *time.Time           { return u.lastLogin }
func (u *User) LockedUntil() *time.Time         { return u.lockedUntil }

// ---------------- Domain behaviors ----------------

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
