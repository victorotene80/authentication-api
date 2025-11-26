package aggregates

import (
    "fmt"
    "time"
    "github.com/google/uuid"
    "authentication/internal/domain/entities"
    "authentication/internal/domain/events"
    "authentication/internal/domain/valueobjects"
)

type UserAggregate struct {
    *AggregateRoot
    User     *entities.User
    Sessions []*entities.Session
}

func NewUserAggregate(
    username valueobjects.Username,
    email valueobjects.Email,
    passwordHash string,
    phone valueobjects.PhoneNumber,
    firstName, lastName string,
    role valueobjects.Role,
) *UserAggregate {
    id := uuid.New().String()
    user := entities.NewUser(username, email, passwordHash, phone, firstName, lastName, role)
    user.ID = id

    aggregate := &UserAggregate{
        AggregateRoot: NewAggregateRoot(id),
        User:          user,
        Sessions:      make([]*entities.Session, 0),
    }

    aggregate.AddEvent(events.NewUserCreatedEvent(
        id,
        username.String(),
        email.String(),
        role.String(),
    ))

    return aggregate
}

func (u *UserAggregate) ChangePassword(oldPasswordHash, newPasswordHash string) error {
    if u.User.PasswordHash != oldPasswordHash {
        return fmt.Errorf("old password does not match")
    }

    u.User.UpdatePassword(newPasswordHash)
    u.IncrementVersion()

    u.AddEvent(events.NewPasswordChangedEvent(u.ID(), u.User.Email.String()))
    return nil
}

func (u *UserAggregate) Login(ipAddress, userAgent string, refreshToken, accessToken string, expiresAt time.Time) *entities.Session {
    u.User.RecordLogin()
    
    session := entities.NewSession(u.ID(), refreshToken, accessToken, ipAddress, userAgent, expiresAt)
    u.Sessions = append(u.Sessions, session)
    u.IncrementVersion()

    u.AddEvent(events.NewUserLoggedInEvent(u.ID(), u.User.Email.String(), ipAddress))
    return session
}

func (u *UserAggregate) Logout(sessionID string) error {
    for _, session := range u.Sessions {
        if session.ID == sessionID {
            session.Revoke()
            u.IncrementVersion()
            u.AddEvent(events.NewUserLoggedOutEvent(u.ID(), u.User.Email.String()))
            return nil
        }
    }
    return fmt.Errorf("session not found")
}

func (u *UserAggregate) RevokeAllSessions() {
    for _, session := range u.Sessions {
        if session.IsValid() {
            session.Revoke()
        }
    }
    u.IncrementVersion()
}

func (u *UserAggregate) UpdateProfile(firstName, lastName string, phone valueobjects.PhoneNumber) {
    u.User.UpdateProfile(firstName, lastName, phone)
    u.IncrementVersion()
    u.AddEvent(events.NewUserUpdatedEvent(u.ID(), u.User.Email.String()))
}

func (u *UserAggregate) Deactivate() {
    u.User.Deactivate()
    u.RevokeAllSessions()
    u.IncrementVersion()
}

func (u *UserAggregate) Activate() {
    u.User.Activate()
    u.IncrementVersion()
}