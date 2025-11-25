package admin

type AdminLockUserRequest struct {
	UserID   string `json:"user_id" validate:"required,uuid"`
	Reason   string `json:"reason" validate:"required,min=10,max=500"`
	Duration int    `json:"duration_minutes,omitempty" validate:"omitempty,min=1"`
}