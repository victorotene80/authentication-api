package events

type UserLoggedInPayload struct {
    UserID    string `json:"user_id"`
    Email     string `json:"email"`
    IPAddress string `json:"ip_address"`
}

func NewUserLoggedInEvent(userID, email, ip string) DomainEvent {
    return newEvent(
        "user.logged_in",
        userID,
        UserLoggedInPayload{UserID: userID, Email: email, IPAddress: ip},
        nil,
    )
}
