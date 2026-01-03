package user

import (
	"errors"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/kernel"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type User struct {
	id           kernel.UserID
	email        string
	passwordHash string
	status       UserStatus
}

type UserStatus string

const (
	UserStatusActive   UserStatus = "ACTIVE"
	UserStatusDisabled UserStatus = "DISABLED"
)

func NewUser(id kernel.UserID, email, passwordHash string, status UserStatus) *User {
	return &User{
		id:           id,
		email:        email,
		passwordHash: passwordHash,
		status:       status,
	}
}

func NewUserFromRegister(id kernel.UserID, email, passwordHash string) *User {
	return &User{
		id:           id,
		email:        email,
		passwordHash: passwordHash,
		status:       UserStatusActive,
	}
}

func (u *User) ID() kernel.UserID {
	return u.id
}

func (u *User) Email() string {
	return u.email
}

func (u *User) PasswordHash() string {
	return u.passwordHash
}

func (u *User) Status() string {
	return string(u.status)
}

func (u *User) SetPasswordHash(hash string) {
	u.passwordHash = hash
}
