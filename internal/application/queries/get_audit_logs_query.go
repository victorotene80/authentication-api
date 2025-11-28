package queries

import "time"

type GetAuditLogsQuery struct {
    UserID       *string
    Action       *string
    ResourceType *string
    ResourceID   *string
    Status       *string
    StartDate    *time.Time
    EndDate      *time.Time
    Page         int
    PageSize     int
}

func (q GetAuditLogsQuery) QueryName() string {
	return "GetAuditLogsQuery"
}