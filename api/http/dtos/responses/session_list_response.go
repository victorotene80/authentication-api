package responses

type SessionListResponse struct {
	Sessions []SessionResponse `json:"sessions"`
	Total    int               `json:"total"`
	Current  *SessionResponse  `json:"current,omitempty"`
}