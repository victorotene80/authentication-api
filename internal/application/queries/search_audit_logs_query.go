package queries

import "time"

type SearchAuditLogsQuery struct {
    Query     string
    StartDate *time.Time
    EndDate   *time.Time
    Page      int
    PageSize  int
}

func (q SearchAuditLogsQuery) QueryName() string {
	return "SearchAuditLogsQuery"
}