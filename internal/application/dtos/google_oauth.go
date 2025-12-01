package dtos

type GoogleOAuthResponse struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

/*func (g *GoogleOAuthResponse) ToRegistrationRequest() OAuthRegistrationRequest {
	return OAuthRegistrationRequest{
		OAuthProvider: "google",
		OAuthID:       g.Sub,
		Email:         g.Email,
		FirstName:     g.GivenName,
		LastName:      g.FamilyName,
		Picture:       g.Picture,
	}
}*/