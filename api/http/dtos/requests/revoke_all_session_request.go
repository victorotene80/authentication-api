package requests

type RevokeAllSessionsRequest struct {
	ExceptCurrent bool `json:"except_current"`
}