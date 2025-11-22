package dto

type RegisterUserRequest struct {
    Email     string `json:"email" binding:"required,email"`
    Username  string `json:"username" binding:"required,min=3,max=30,alphanumunicode"`
    Password  string `json:"password" binding:"required,min=12,max=128"`
    FirstName string `json:"first_name" binding:"required,min=1,max=100"`
    LastName  string `json:"last_name" binding:"required,min=1,max=100"`
    //Phone     string `json:"phone,omitempty" binding:"omitempty,e164"` // optional, use custom validator for e164 if needed
}

type RegisterUserResponse struct {
    Message string `json:"message"`
}
