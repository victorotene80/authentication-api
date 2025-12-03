package response

type LoginResponse struct {
	User         UserInfo `json:"user"`
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int64    `json:"expires_in"` 
}

type UserInfo struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Username  string `json:"username,omitempty"` 
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
	IsVerified bool  `json:"is_verified"`
}
