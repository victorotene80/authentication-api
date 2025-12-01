package request

/*import (
	"fmt"

	"github.com/go-playground/validator/v10"

	"authentication/internal/application/commands"
)

type LoginUserRequest struct {
    Email    string `json:"email" validate:"required"`
    Password string `json:"password" validate:"required"`
}


func (r *LoginUserRequest) ToCommand(ipAddress, userAgent string) commands.LoginUserCommand {
	return commands.LoginUserCommand{
		Email:     r.Email,
		Password:  r.Password,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
}

func (r *LoginUserRequest) Validate(v *validator.Validate) error {
	if v == nil {
		return fmt.Errorf("validator is nil")
	}
	return v.Struct(r)
}*/
