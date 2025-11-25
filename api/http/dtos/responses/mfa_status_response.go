package responses

type MFAStatusResponse struct {
	Enabled             bool `json:"enabled"`
	Method              string `json:"method,omitempty"`
	BackupCodesRemaining int `json:"backup_codes_remaining,omitempty"`
}