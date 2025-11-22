package events

import (
	"authentication/shared/utils"

	"github.com/google/uuid"
)


type UserCreatedEvent struct {
	BaseDomainEvent
}

func NewUserCreatedEvent(userID uuid.UUID, email string, username string) *UserCreatedEvent {
	return &UserCreatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          uuid.New(),
			Name:        "UserCreated",
			Timestamp:   utils.NowUTC(),
			AggregateId: userID,
			Data: map[string]interface{}{
				"user_id":  userID.String(),
				"email":    email,
				"username": username,
			},
		},
	}
}
