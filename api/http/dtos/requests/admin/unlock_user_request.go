package admin

type AdminUnlockUserRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
	Reason string `json:"reason" validate:"required,min=10,max=500"`
}