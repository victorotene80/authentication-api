package services

import "context"

type OAuthUserInfo struct {
	OAuthID       string
	Email         string
	EmailVerified bool
	FirstName     string
	LastName      string
	FullName      string
	AvatarURL     string
}

type OAuthService interface {
	Verify(ctx context.Context, provider, idToken, accessToken string) (*OAuthUserInfo, error)
}
