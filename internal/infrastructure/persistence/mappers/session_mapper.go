package mappers

import (
	"authentication/internal/domain/entities"
	"authentication/internal/infrastructure/persistence/database/models"
)

type SessionMapper struct{}

func NewSessionMapper() *SessionMapper {
	return &SessionMapper{}
}

func (m *SessionMapper) ToModel(session *entities.Session) *models.SessionModel {
	return &models.SessionModel{
		ID:           session.ID,
		UserID:       session.UserID,
		RefreshToken: session.RefreshToken,
		AccessToken:  session.AccessToken,
		IPAddress:    session.IPAddress,
		UserAgent:    session.UserAgent,
		ExpiresAt:    session.ExpiresAt,
		IsRevoked:    session.IsRevoked,
		RevokedAt:    session.RevokedAt,
		CreatedAt:    session.CreatedAt,
		UpdatedAt:    session.UpdatedAt,
	}
}

func (m *SessionMapper) ToDomain(model *models.SessionModel) *entities.Session {
	return &entities.Session{
		ID:           model.ID,
		UserID:       model.UserID,
		RefreshToken: model.RefreshToken,
		AccessToken:  model.AccessToken,
		IPAddress:    model.IPAddress,
		UserAgent:    model.UserAgent,
		ExpiresAt:    model.ExpiresAt,
		IsRevoked:    model.IsRevoked,
		RevokedAt:    model.RevokedAt,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}
