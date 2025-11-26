package aggregates

import (
	"github.com/google/uuid"
)

type OAuthClient struct {
	*AggregateRoot
	ClientID     string
	ClientSecret string
	Provider     string
	Name         string
	RedirectURIs []string
	Scopes       []string
	IsActive     bool
}

func NewOAuthClient(
	clientID string,
	clientSecret string,
	provider string,
	name string,
	redirectURIs []string,
	scopes []string,
) *OAuthClient {
	id := uuid.New().String()
	return &OAuthClient{
		AggregateRoot: NewAggregateRoot(id),
		ClientID:      clientID,
		ClientSecret:  clientSecret,
		Provider:      provider,
		Name:          name,
		RedirectURIs:  redirectURIs,
		Scopes:        scopes,
		IsActive:      true,
	}
}

func (o *OAuthClient) Activate() {
	o.IsActive = true
	o.IncrementVersion()
}

func (o *OAuthClient) Deactivate() {
	o.IsActive = false
	o.IncrementVersion()
}

func (o *OAuthClient) UpdateScopes(scopes []string) {
	o.Scopes = scopes
	o.IncrementVersion()
}

func (o *OAuthClient) UpdateRedirectURIs(redirectURIs []string) {
	o.RedirectURIs = redirectURIs
	o.IncrementVersion()
}

func (o *OAuthClient) RotateSecret(newSecret string) {
	o.ClientSecret = newSecret
	o.IncrementVersion()
}
