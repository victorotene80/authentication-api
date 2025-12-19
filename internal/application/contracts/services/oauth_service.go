// File: internal/application/contracts/services/oauth_service.go
package services

import "context"

// OAuthUserInfo contains verified information from the OAuth provider
type OAuthUserInfo struct {
	ProviderUserID string // ID from OAuth provider (e.g., Google's user ID)
	Email          string
	EmailVerified  bool
	FirstName      string
	LastName       string
	ProfilePicture string
	Locale         string
}

// OAuthProvider represents supported OAuth providers
type OAuthProvider string

const (
	OAuthProviderGoogle OAuthProvider = "google"
	OAuthProviderGithub OAuthProvider = "github"
	// Add more as needed
)

// OAuthService handles OAuth provider interactions
// This is an APPLICATION service - it coordinates with external OAuth providers
type OAuthService interface {
	// Verify validates the OAuth tokens and returns user information
	// Parameters:
	//   - provider: The OAuth provider (google, github, etc.)
	//   - idToken: The ID token from the OAuth provider (if applicable)
	//   - accessToken: The access token from the OAuth provider
	// Returns verified user information from the OAuth provider
	Verify(ctx context.Context, provider, idToken, accessToken string) (*OAuthUserInfo, error)

	// GetAuthorizationURL generates the OAuth authorization URL for the given provider
	// Used when initiating the OAuth flow
	GetAuthorizationURL(ctx context.Context, provider string, state string) (string, error)

	// ExchangeCodeForTokens exchanges an authorization code for access tokens
	// Used in the OAuth callback after user authorizes
	ExchangeCodeForTokens(ctx context.Context, provider, code string) (idToken, accessToken string, err error)

	// RevokeAccess revokes the user's access token with the OAuth provider
	// Used when user disconnects their OAuth account
	RevokeAccess(ctx context.Context, provider, accessToken string) error

	// ValidateProvider checks if the provider is supported
	ValidateProvider(provider string) error
}

// OAuthServiceImpl would be implemented in infrastructure layer
// Example structure (don't include in interface file):
//
// type OAuthServiceImpl struct {
//     googleProvider  *GoogleOAuthProvider
//     githubProvider  *GithubOAuthProvider
//     logger          logging.Logger
// }