package events

type PasswordChangedPayload struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
}

func NewPasswordChangedEvent(userID, email string) DomainEvent {
    return newEvent(
        "user.password_changed",
        userID,
        PasswordChangedPayload{UserID: userID, Email: email},
        nil,
    )
}
