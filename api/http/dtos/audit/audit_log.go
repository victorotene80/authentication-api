package audit

import "time"

type AuditLogDTO struct {
    ID            string                 `json:"id"`
    UserID        string                 `json:"user_id"`
    Action        string                 `json:"action"`
    ResourceType  string                 `json:"resource_type"`
    ResourceID    string                 `json:"resource_id,omitempty"`
    IPAddress     string                 `json:"ip_address"`
    UserAgent     string                 `json:"user_agent"`
    Status        string                 `json:"status"`
    ErrorMessage  string                 `json:"error_message,omitempty"`
    Metadata      map[string]interface{} `json:"metadata,omitempty"`
    Timestamp     time.Time              `json:"timestamp"`
}