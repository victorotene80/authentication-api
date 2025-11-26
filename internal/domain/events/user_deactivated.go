package events

type UserDeactivatedPayload struct {
    UserID string `json:"user_id"`
}

func NewUserDeactivatedEvent(userID string) DomainEvent {
    return newEvent(
        "user.deactivated",
        userID,
        UserDeactivatedPayload{UserID: userID},
        nil,
    )
}
