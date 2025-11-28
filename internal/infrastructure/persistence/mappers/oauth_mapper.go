package mappers

import (
    "encoding/json"
    "authentication/internal/domain/aggregates"
    "authentication/internal/domain/entities"
    "authentication/internal/infrastructure/persistence/database/models"
)

type OAuthMapper struct{}

func NewOAuthMapper() *OAuthMapper {
    return &OAuthMapper{}
}

func (m *OAuthMapper) ClientToModel(aggregate *aggregates.OAuthClient) (*models.OAuthClientModel, error) {
    redirectURIsJSON, err := json.Marshal(aggregate.RedirectURIs)
    if err != nil {
        return nil, err
    }

    scopesJSON, err := json.Marshal(aggregate.Scopes)
    if err != nil {
        return nil, err
    }

    return &models.OAuthClientModel{
        ID:           aggregate.ID(),
        ClientID:     aggregate.ClientID,
        ClientSecret: aggregate.ClientSecret,
        Provider:     aggregate.Provider,
        Name:         aggregate.Name,
        RedirectURIs: redirectURIsJSON,
        Scopes:       scopesJSON,
        IsActive:     aggregate.IsActive,
        Version:      aggregate.Version(),
    }, nil
}

func (m *OAuthMapper) ClientToDomain(model *models.OAuthClientModel) (*aggregates.OAuthClient, error) {
    var redirectURIs []string
    if err := json.Unmarshal(model.RedirectURIs, &redirectURIs); err != nil {
        return nil, err
    }

    var scopes []string
    if err := json.Unmarshal(model.Scopes, &scopes); err != nil {
        return nil, err
    }

    aggregate := &aggregates.OAuthClient{
        AggregateRoot: aggregates.NewAggregateRoot(model.ID),
        ClientID:      model.ClientID,
        ClientSecret:  model.ClientSecret,
        Provider:      model.Provider,
        Name:          model.Name,
        RedirectURIs:  redirectURIs,
        Scopes:        scopes,
        IsActive:      model.IsActive,
    }

    return aggregate, nil
}

func (m *OAuthMapper) TokenToModel(token *entities.OAuthToken) *models.OAuthTokenModel {
    return &models.OAuthTokenModel{
        ID:           token.ID,
        UserID:       token.UserID,
        Provider:     token.Provider,
        AccessToken:  token.AccessToken,
        RefreshToken: token.RefreshToken,
        TokenType:    token.TokenType,
        ExpiresAt:    token.ExpiresAt,
        Scope:        token.Scope,
        IsRevoked:    token.IsRevoked,
        RevokedAt:    token.RevokedAt,
        CreatedAt:    token.CreatedAt,
        UpdatedAt:    token.UpdatedAt,
    }
}

func (m *OAuthMapper) TokenToDomain(model *models.OAuthTokenModel) *entities.OAuthToken {
    return &entities.OAuthToken{
        ID:           model.ID,
        UserID:       model.UserID,
        Provider:     model.Provider,
        AccessToken:  model.AccessToken,
        RefreshToken: model.RefreshToken,
        TokenType:    model.TokenType,
        ExpiresAt:    model.ExpiresAt,
        Scope:        model.Scope,
        IsRevoked:    model.IsRevoked,
        RevokedAt:    model.RevokedAt,
        CreatedAt:    model.CreatedAt,
        UpdatedAt:    model.UpdatedAt,
    }
}