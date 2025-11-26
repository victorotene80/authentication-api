package request

import "time"

type SearchAuditLogsRequest struct {
	Query     string     `json:"query"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Page      int        `json:"page" validate:"min=1"`
	PageSize  int        `json:"page_size" validate:"min=1,max=100"`
}
