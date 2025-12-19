package events

type UserCreatedPayload struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Type      string `json:"type"`
	Status    string `json:"status"`
}

func NewUserCreatedEvent(
	userID, username, email, role, firstName, lastName string,
	ipAddress, userAgent, status, registrationType string, //deviceID string,
) DomainEvent {
	payload := UserCreatedPayload{
		UserID:    userID,
		Username:  username,
		Email:     email,
		Role:      role,
		FirstName: firstName,
		LastName:  lastName,
		Type:      registrationType,
		Status:    status,
	}

	return newEvent(
		"user.created",
		userID,
		payload,
		map[string]string{
			"event_source": "auth-service",
			"ip_address":   ipAddress,
			"user_agent":   userAgent,
			"status":       status,
			//"deviceID":     deviceID,
		},
	)
}

/*func NewUserCreatedEvent(
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
}*/
