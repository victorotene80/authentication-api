package response

type OAuthLoginResponse struct {
	LoginResponse
	IsNewUser bool `json:"is_new_user"` 
}