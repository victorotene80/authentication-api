package valueobjects

type Password struct {
	value string
}

func NewPassword(hashed string) Password {
	return Password{value: hashed}
}

func (p Password) Value() string {
	return p.value
}

func (p Password) IsEmpty() bool {
	return p.value == ""
}
