package user

import (
	"context"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/user"
)

type LoginUserUsecase struct {
	userRepository user.UserRepository
	userService    *user.UserService
	tokenService   user.TokenService
}

func NewLoginUserUsecase(
	userRepository user.UserRepository,
	userService *user.UserService,
	tokenService user.TokenService,
) *LoginUserUsecase {
	return &LoginUserUsecase{
		userRepository: userRepository,
		userService:    userService,
		tokenService:   tokenService,
	}
}

// Execute handles user login by verifying the user's credentials and generating a JWT token.
// It retrieves the user by email, verifies the password, and generates a JWT token if successful.
// It returns the user information and the generated token, or an error if any step fails.
func (uc *LoginUserUsecase) Execute(ctx context.Context, command user.LoginUserCommand) (*user.User, string, error) {
	// Retrieve the user by email
	userEntity, err := uc.userRepository.GetUserByEmail(ctx, command.Email)
	if err != nil {
		return nil, "", &user.RepositoryError{
			Operation: "get_user_by_email",
			Wrapped:   err,
		}
	}

	if userEntity == nil {
		return nil, "", &user.UserNotFoundError{Email: command.Email, UserId: ""}
	}

	// Verify the password
	err = uc.userService.VerifyPassword(userEntity.PasswordHash, command.Password)
	if err != nil {
		return nil, "", user.ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := uc.tokenService.GenerateToken(userEntity)
	if err != nil {
		return nil, "", &user.RepositoryError{
			Operation: "generate_token",
			Wrapped:   err,
		}
	}

	return userEntity, token, nil
}
