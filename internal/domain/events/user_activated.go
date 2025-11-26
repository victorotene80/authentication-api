package events

type UserActivatedPayload struct {
    UserID string `json:"user_id"`
}

func NewUserActivatedEvent(userID string) DomainEvent {
    return newEvent(
        "user.activated",
        userID,
        UserActivatedPayload{UserID: userID},
        nil,
    )
}
