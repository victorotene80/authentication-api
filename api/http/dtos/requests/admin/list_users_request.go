package admin

type AdminListUsersRequest struct {
	Page     int    `json:"page" validate:"min=1"`
	PageSize int    `json:"page_size" validate:"min=1,max=100"`
	Status   string `json:"status,omitempty" validate:"omitempty,oneof=active inactive locked pending"`
	Search   string `json:"search,omitempty"`
	SortBy   string `json:"sort_by,omitempty" validate:"omitempty,oneof=created_at email username last_login"`
	SortDir  string `json:"sort_dir,omitempty" validate:"omitempty,oneof=asc desc"`
}