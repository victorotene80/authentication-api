package repositories

import (
	"context"
	//"time"
	"authentication/internal/domain/aggregates"
	//"authentication/internal/domain/valueobjects"
)

/*type AuditLogFilter struct {
	UserID       *string
	Action       *valueobjects.AuditAction
	ResourceType *string
	ResourceID   *string
	Status       *string
	StartDate    *time.Time
	EndDate      *time.Time
	Page         int
	PageSize     int
}*/

type AuditRepository interface {
	Create(ctx context.Context, log *aggregates.AuditLog) error
	//List(ctx context.Context, filter AuditLogFilter) ([]*aggregates.AuditLog, error)
}
 