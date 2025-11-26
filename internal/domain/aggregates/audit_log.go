package aggregates

import (
	"authentication/internal/domain/valueobjects"
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	*AggregateRoot
	UserID       string
	Action       valueobjects.AuditAction
	ResourceType string
	ResourceID   string
	IPAddress    string
	UserAgent    string
	Status       string
	ErrorMessage string
	Metadata     map[string]interface{}
	Timestamp    time.Time
}

func NewAuditLog(
	userID string,
	action valueobjects.AuditAction,
	resourceType string,
	resourceID string,
	ipAddress string,
	userAgent string,
	status string,
	metadata map[string]interface{},
) *AuditLog {
	id := uuid.New().String()
	return &AuditLog{
		AggregateRoot: NewAggregateRoot(id),
		UserID:        userID,
		Action:        action,
		ResourceType:  resourceType,
		ResourceID:    resourceID,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		Status:        status,
		Metadata:      metadata,
		Timestamp:     time.Now(),
	}
}

func NewAuditLogWithError(
	userID string,
	action valueobjects.AuditAction,
	resourceType string,
	resourceID string,
	ipAddress string,
	userAgent string,
	errorMessage string,
	metadata map[string]interface{},
) *AuditLog {
	log := NewAuditLog(userID, action, resourceType, resourceID, ipAddress, userAgent, "FAILURE", metadata)
	log.ErrorMessage = errorMessage
	return log
}
