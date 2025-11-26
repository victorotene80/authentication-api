package events

type UserUpdatedPayload struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
}

func NewUserUpdatedEvent(userID, email string) DomainEvent {
    return newEvent(
        "user.updated",
        userID,
        UserUpdatedPayload{UserID: userID, Email: email},
        nil,
    )
}
