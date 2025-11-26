package valueobjects

import "fmt"

type Role string

const (
    RoleUser  Role = "USER"
    RoleAdmin Role = "ADMIN"
    RoleModerator Role = "MODERATOR"
)

func NewRole(role string) (Role, error) {
    r := Role(role)
    if !r.IsValid() {
        return "", fmt.Errorf("invalid role: %s", role)
    }
    return r, nil
}

func (r Role) IsValid() bool {
    switch r {
    case RoleUser, RoleAdmin, RoleModerator:
        return true
    }
    return false
}

func (r Role) String() string {
    return string(r)
}

func (r Role) HasPermission(required Role) bool {
    if r == RoleAdmin {
        return true
    }
    if r == RoleModerator && required == RoleUser {
        return true
    }
    return r == required
}