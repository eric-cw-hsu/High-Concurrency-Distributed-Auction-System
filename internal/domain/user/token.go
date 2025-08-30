package user

type TokenService interface {
	GenerateToken(user *User) (string, error)
	VerifyToken(token string) (*User, error)
}
