package user

import (
	"context"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/user"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/usecase/wallet"
	"github.com/samborkent/uuidv7"
)

type RegisterUserUsecase struct {
	userRepository user.UserRepository
	userService    *user.UserService
	walletService  wallet.WalletService
}

func NewRegisterUserUsecase(
	userRepository user.UserRepository,
	userService *user.UserService,
	walletService wallet.WalletService,
) *RegisterUserUsecase {
	return &RegisterUserUsecase{
		userRepository: userRepository,
		userService:    userService,
		walletService:  walletService,
	}
}

// Execute handles user registration by creating a new user and associated wallet.
// It hashes the password, creates the user, creates a wallet for the user, and returns the created user.
func (uc *RegisterUserUsecase) Execute(ctx context.Context, command user.RegisterUserCommand) (*user.User, error) {
	hashedPassword, err := uc.userService.HashPassword(command.Password)
	if err != nil {
		return nil, err
	}

	userEntity := user.User{
		ID:           uuidv7.New().String(),
		Email:        command.Email,
		Name:         command.Name,
		PasswordHash: hashedPassword,
	}

	// Check if user already exists
	existingUser, err := uc.userRepository.GetUserByEmail(ctx, userEntity.Email)
	if err != nil {
		return nil, &user.RepositoryError{
			Operation: "get_user_by_email",
			Wrapped:   err,
		}
	}

	if existingUser != nil {
		return nil, &user.UserAlreadyExistsError{
			Email: userEntity.Email,
		}
	}

	// Create the user
	createdUser, err := uc.userRepository.SaveUser(ctx, &userEntity)
	if err != nil {
		return nil, &user.RepositoryError{
			Operation: "save_user",
			Wrapped:   err,
		}
	}

	// Create wallet for the user
	_, err = uc.walletService.EnsureWalletExists(ctx, createdUser.ID)
	if err != nil {
		// Log the error but don't fail user creation
		// In a production system, you might want to handle this differently
		// or use saga pattern for distributed transactions
		// We could return a warning or use a separate error type
		return createdUser, &user.RepositoryError{
			Operation: "create_wallet",
			Wrapped:   err,
		}
	}

	return createdUser, nil
}
