package events

type UserCreatedPayload struct {
    UserID   string `json:"user_id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Role     string `json:"role"`
}

func NewUserCreatedEvent(
    userID, username, email, role string,
) DomainEvent {

    payload := UserCreatedPayload{
        UserID:   userID,
        Username: username,
        Email:    email,
        Role:     role,
    }

    return newEvent(
        "user.created",
        userID,
        payload,
        map[string]string{
            "event_source": "auth-service",
        },
    )
}
