package kernel

import "errors"

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

func NewRole(raw string) (Role, error) {
	switch raw {
	case string(RoleAdmin), string(RoleUser):
		return Role(raw), nil
	default:
		return "", errors.New("invalid role")
	}
}

func (r Role) String() string {
	return string(r)
}
