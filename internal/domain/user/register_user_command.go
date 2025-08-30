package user

type RegisterUserCommand struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8,max=50"` // Password must be at least 8 characters long
}
