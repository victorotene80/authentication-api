package requests

type UpdateProfileRequest struct {
	FirstName string `json:"first_name,omitempty" validate:"omitempty,min=1,max=50"`
	LastName  string `json:"last_name,omitempty" validate:"omitempty,min=1,max=50"`
	Phone     string `json:"phone,omitempty" validate:"omitempty,e164"`
}