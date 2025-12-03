package aggregates

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"authentication/internal/domain/entities"
	"authentication/internal/domain/events"
	"authentication/internal/domain/services"
	"authentication/internal/domain/valueobjects"
)

type UserAggregate struct {
	*AggregateRoot
	User     *entities.User
	Sessions []*entities.Session
}

func NewEmailUserAggregate(
    username valueobjects.Username,
    email valueobjects.Email,
    password valueobjects.Password,
    //phone valueobjects.PhoneNumber,
    firstName, lastName string,
    role valueobjects.Role,
) *UserAggregate {

    id := uuid.New().String()
    user := entities.NewUser(username, email, password, firstName, lastName, role, false, "")
    user.ID = id

    agg := &UserAggregate{
        AggregateRoot: NewAggregateRoot(id),
        User:          user,
        Sessions:      []*entities.Session{},
    }

    agg.AddEvent(events.NewUserCreatedEvent(
        id, username.String(), email.String(), role.String(),
    ))

    return agg
}

func NewOAuthUserAggregate(
    email valueobjects.Email,
	firstName, lastName string,
    oauthProvider string,
    role valueobjects.Role,
) *UserAggregate {

    id := uuid.New().String()

	emptyUsername := valueobjects.EmptyUsername()
    emptyPassword := valueobjects.EmptyPassword()

    user := entities.NewUser(emptyUsername, email, emptyPassword, firstName, lastName, role, true, oauthProvider)
    user.ID = id
    user.VerifyEmail()

    agg := &UserAggregate{
        AggregateRoot: NewAggregateRoot(id),
        User:          user,
        Sessions:      []*entities.Session{},
    }

    agg.AddEvent(events.NewUserCreatedEvent(
        id, "", email.String(), role.String(),
    ))

    return agg
}

func (u *UserAggregate) ChangePassword(
	oldPlainPassword string,
	newPlainPassword string,
	hasher *services.PasswordHashingService,
) error {
	if !hasher.Verify(oldPlainPassword, u.User.Password) {
		return fmt.Errorf("old password does not match")
	}

	newPasswordVO, err := hasher.HashPassword(newPlainPassword)
	if err != nil {
		return err
	}

	u.User.UpdatePassword(newPasswordVO)

	u.IncrementVersion()
	u.AddEvent(events.NewPasswordChangedEvent(u.ID(), u.User.Email.String()))
	return nil
}

func (u *UserAggregate) Login(
	ipAddress, userAgent string,
	refreshToken, accessToken string,
	expiresAt time.Time,
) *entities.Session {
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
