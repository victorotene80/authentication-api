package request

import(
	"github.com/go-playground/validator/v10"
	"authentication/internal/application/commands"
)

type EmailLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func (r *EmailLoginRequest) Validate(v *validator.Validate) error {
	return v.Struct(r)
}

func (r *EmailLoginRequest) ToCommand(ip, ua string) commands.LoginEmailUserCommand {
	return commands.LoginEmailUserCommand{
		Email:     r.Email,
		Password:  r.Password,
		IPAddress: ip,
		UserAgent: ua,
	}
}