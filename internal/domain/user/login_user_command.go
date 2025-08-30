package user

type LoginUserCommand struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=50"` // Password must be at least 8 characters long
}
