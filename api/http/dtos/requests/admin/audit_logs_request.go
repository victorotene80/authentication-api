package admin

type AdminGetAuditLogsRequest struct {
	UserID    string `json:"user_id,omitempty" validate:"omitempty,uuid"`
	Action    string `json:"action,omitempty"`
	StartDate string `json:"start_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
	EndDate   string `json:"end_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
	Page      int    `json:"page" validate:"min=1"`
	PageSize  int    `json:"page_size" validate:"min=1,max=100"`
}