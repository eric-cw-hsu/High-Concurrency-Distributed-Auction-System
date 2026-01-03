package kernel

import "errors"

type UserID string

var ErrUserIDCannotBeEmpty = errors.New("user ID cannot be empty")

func NewUserID(raw string) (UserID, error) {
	if raw == "" {
		return "", ErrUserIDCannotBeEmpty
	}

	return UserID(raw), nil
}

func (u UserID) String() string {
	return string(u)
}
