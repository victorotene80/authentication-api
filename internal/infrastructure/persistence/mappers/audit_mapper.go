package mappers

import (
	"authentication/internal/domain/aggregates"
	"authentication/internal/domain/valueobjects"
	"authentication/internal/infrastructure/persistence/database/models"
	"encoding/json"
)

type AuditMapper struct{}

func NewAuditMapper() *AuditMapper {
	return &AuditMapper{}
}

func (m *AuditMapper) ToModel(aggregate *aggregates.AuditLog) (*models.AuditLogModel, error) {
	var metadataJSON []byte
	var err error

	if aggregate.Metadata != nil {
		metadataJSON, err = json.Marshal(aggregate.Metadata)
		if err != nil {
			return nil, err
		}
	}

	return &models.AuditLogModel{
		ID:           aggregate.ID(),
		UserID:       aggregate.UserID,
		Action:       aggregate.Action.String(),
		ResourceType: aggregate.ResourceType,
		ResourceID:   aggregate.ResourceID,
		IPAddress:    aggregate.IPAddress,
		UserAgent:    aggregate.UserAgent,
		Status:       aggregate.Status,
		ErrorMessage: aggregate.ErrorMessage,
		Metadata:     metadataJSON,
		Timestamp:    aggregate.Timestamp,
	}, nil
}

func (m *AuditMapper) ToDomain(model *models.AuditLogModel) (*aggregates.AuditLog, error) {
	var metadata map[string]interface{}
	if len(model.Metadata) > 0 {
		if err := json.Unmarshal(model.Metadata, &metadata); err != nil {
			return nil, err
		}
	}

	aggregate := &aggregates.AuditLog{
		AggregateRoot: aggregates.NewAggregateRoot(model.ID),
		UserID:        model.UserID,
		Action:        valueobjects.AuditAction(model.Action),
		ResourceType:  model.ResourceType,
		ResourceID:    model.ResourceID,
		IPAddress:     model.IPAddress,
		UserAgent:     model.UserAgent,
		Status:        model.Status,
		ErrorMessage:  model.ErrorMessage,
		Metadata:      metadata,
		Timestamp:     model.Timestamp,
	}

	return aggregate, nil
}
