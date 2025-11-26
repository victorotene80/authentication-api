package response

import audit "authentication/api/http/dtos/audit"

type GetAuditLogsResponse struct {
	Logs       []audit.AuditLogDTO `json:"logs"`
	TotalCount int64               `json:"total_count"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
}
