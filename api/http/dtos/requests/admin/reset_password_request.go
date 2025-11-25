package admin

type AdminResetPasswordRequest struct {
	UserID            string `json:"user_id" validate:"required,uuid"`
	NotifyUser        bool   `json:"notify_user"`
	TemporaryPassword string `json:"temporary_password,omitempty"`
}