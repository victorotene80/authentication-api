package requests

type DeleteAccountRequest struct {
	Password string `json:"password" validate:"required"`
	Reason   string `json:"reason,omitempty" validate:"omitempty,max=500"`
	Confirm  bool   `json:"confirm" validate:"required,eq=true"`
}