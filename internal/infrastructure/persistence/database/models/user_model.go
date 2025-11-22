package models

import(
	"time"
	"github.com/google/uuid"
)

type UserModel struct {
    ID                   uuid.UUID
    Email                string
    Username             string
    PasswordHash         string
    FirstName            string
    LastName             string
    Phone                *string
    EmailVerified        bool
    EmailVerifiedAt      *time.Time
    Status               string
    IsLocked             bool
    LockedUntil          *time.Time
    FailedLoginAttempts  int
    LastFailedLogin      *time.Time
    TwoFactorEnabled     bool
    TwoFactorSecret      *string
    CreatedAt            time.Time
    UpdatedAt            time.Time
    LastLogin            *time.Time
}