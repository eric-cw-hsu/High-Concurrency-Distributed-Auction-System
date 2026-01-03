package service

type PasswordVerifier interface {
	Verify(hash string, plain string) bool
	Hash(plain string) (string, error)
}
