package request

import "time"

type GetAuditLogsRequest struct {
    UserID       string     `json:"user_id,omitempty"`
    Action       string     `json:"action,omitempty"`
    ResourceType string     `json:"resource_type,omitempty"`
    StartDate    *time.Time `json:"start_date,omitempty"`
    EndDate      *time.Time `json:"end_date,omitempty"`
    Page         int        `json:"page" validate:"min=1"`
    PageSize     int        `json:"page_size" validate:"min=1,max=100"`
}