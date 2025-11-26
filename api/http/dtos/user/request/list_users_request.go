package user

type ListUsersRequest struct {
    Page     int    `json:"page" validate:"min=1"`
    PageSize int    `json:"page_size" validate:"min=1,max=100"`
    Role     string `json:"role,omitempty"`
    IsActive *bool  `json:"is_active,omitempty"`
}