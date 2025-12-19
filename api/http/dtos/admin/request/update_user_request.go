package user

type UpdateUserRequest struct {
    FirstName string `json:"first_name,omitempty"`
    LastName  string `json:"last_name,omitempty"`
    Phone     string `json:"phone,omitempty" validate:"omitempty,e164"`
}