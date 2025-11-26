package repositories

import (
    "context"
    "time"
    "authentication/internal/domain/aggregates"
    "authentication/internal/domain/valueobjects"
)

type AuditLogFilter struct {
    UserID       *string
    Action       *valueobjects.AuditAction
    ResourceType *string
    ResourceID   *string
    Status       *string
    StartDate    *time.Time
    EndDate      *time.Time
    Page         int
    PageSize     int
}

type AuditRepository interface {
    Create(ctx context.Context, log *aggregates.AuditLog) error
    FindByID(ctx context.Context, id string) (*aggregates.AuditLog, error)
    FindByUserID(ctx context.Context, userID string, page, pageSize int) ([]*aggregates.AuditLog, int64, error)
    Search(ctx context.Context, filter *AuditLogFilter) ([]*aggregates.AuditLog, int64, error)
    List(ctx context.Context, page, pageSize int) ([]*aggregates.AuditLog, int64, error)
    DeleteOlderThan(ctx context.Context, date time.Time) error
}