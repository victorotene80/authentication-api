package request

import(
	"github.com/go-playground/validator/v10"
)

type LoginEmailUserCommand struct {
	Email     string
	Password  string
	IPAddress string
	UserAgent string
	DeviceID  string
}

func (r *LoginEmailUserCommand) Validate(v *validator.Validate) error {
	return v.Struct(r)
}

/*func(r *LoginEmailUserCommand) ToCommand(ip, ua string) LoginEmailUserCommand {
	return LoginEmailUserCommand{
		Email:     r.Email,
		Password:  r.Password,
		IPAddress: ip,
		UserAgent: ua,
	}
}*/
