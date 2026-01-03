package crypto

import (
	domainSvc "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/domain/service"

	"golang.org/x/crypto/bcrypt"
)

type BcryptVerifier struct {
	cost int
}

func NewBcryptVerifier(cost int) domainSvc.PasswordVerifier {
	return &BcryptVerifier{
		cost: cost,
	}
}

func (v *BcryptVerifier) Verify(hash string, plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	return err == nil
}

func (v *BcryptVerifier) Hash(plain string) (string, error) {
	return HashPassword(plain, v.cost)
}

func HashPassword(password string, cost int) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(bytes), err
}
