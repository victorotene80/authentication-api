package admin

import(
	"authentication/api/http/dtos/responses"
)

type AdminUserDetailResponse struct {
	User           AdminUserResponse   `json:"user"`
	Sessions       []responses.SessionResponse   `json:"sessions"`
	RecentActivity []responses.AuditLogResponse  `json:"recent_activity"`
	Statistics     AdminUserStatistics `json:"statistics"`
}

