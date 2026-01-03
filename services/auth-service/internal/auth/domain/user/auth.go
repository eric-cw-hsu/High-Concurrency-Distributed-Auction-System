package user

import "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/domain/service"

func (u *User) Login(password string, verifier service.PasswordVerifier) error {
	if u.status != UserStatusActive {
		return ErrInvalidCredentials
	}

	if !verifier.Verify(u.passwordHash, password) {
		return ErrInvalidCredentials
	}

	return nil
}
