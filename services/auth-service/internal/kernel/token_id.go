package kernel

import "errors"

type TokenID string

var ErrTokenIDCannotBeEmpty = errors.New("token ID cannot be empty")

func NewTokenID(raw string) (TokenID, error) {
	if raw == "" {
		return "", ErrTokenIDCannotBeEmpty
	}
	return TokenID(raw), nil
}

func (t TokenID) String() string {
	return string(t)
}
