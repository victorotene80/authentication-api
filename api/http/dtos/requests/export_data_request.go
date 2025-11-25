package requests

type ExportUserDataRequest struct {
	Format   string   `json:"format" validate:"required,oneof=json csv"`
	Sections []string `json:"sections,omitempty" validate:"omitempty,dive,oneof=profile sessions login_history audit_logs"`
}
