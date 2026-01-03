package user

type UserRepository interface {
	FindByEmail(email string) (*User, error)
	Save(user *User) error
}
