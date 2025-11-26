package request

type RegisterUserRequest struct {
    Username  string `json:"username" validate:"required,min=3,max=50"`
    Email     string `json:"email" validate:"required,email"`
    Password  string `json:"password" validate:"required,min=8"`
    Phone     string `json:"phone,omitempty" validate:"omitempty,e164"`
    FirstName string `json:"first_name" validate:"required"`
    LastName  string `json:"last_name" validate:"required"`
}