package events

type UserLoggedOutPayload struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
}

func NewUserLoggedOutEvent(userID, email string) DomainEvent {
    return newEvent(
        "user.logged_out",
        userID,
        UserLoggedOutPayload{UserID: userID, Email: email},
        nil,
    )
}
