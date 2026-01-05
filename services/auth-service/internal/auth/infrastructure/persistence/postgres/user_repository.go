package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/domain/user"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/common/logger"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/kernel"
	"go.uber.org/zap"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) user.UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	logger.DebugContext(ctx, "querying user by email",
		zap.String("email", email),
	)

	var id, passwordHash, status string

	row := r.db.QueryRowContext(ctx, "SELECT id, password_hash, status FROM users WHERE email = $1", email)
	if err := row.Scan(&id, &passwordHash, &status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.DebugContext(ctx, "user not found",
				zap.String("email", email),
			)
			return nil, errors.New("user not found")
		}

		logger.ErrorContext(ctx, "database query failed",
			zap.String("operation", "FindByEmail"),
			zap.String("email", email),
			zap.Error(err),
		)
		return nil, err
	}

	userID, err := kernel.NewUserID(id)
	if err != nil {
		logger.ErrorContext(ctx, "invalid user id from database",
			zap.String("id", id),
			zap.String("email", email),
			zap.Error(err),
		)
		return nil, err
	}

	u := user.NewUser(
		userID,
		email,
		passwordHash,
		user.UserStatus(status),
	)

	return u, nil
}

func (r *UserRepository) Save(ctx context.Context, u *user.User) error {
	logger.DebugContext(ctx, "saving user",
		zap.String("user_id", string(u.ID())),
		zap.String("email", u.Email()),
	)

	_, err := r.db.ExecContext(
		ctx,
		"INSERT INTO users (id, email, password_hash, status) VALUES ($1, $2, $3, $4)",
		u.ID(),
		u.Email(),
		u.PasswordHash(),
		u.Status(),
	)
	if err != nil {
		logger.ErrorContext(ctx, "failed to save user",
			zap.String("user_id", string(u.ID())),
			zap.String("email", u.Email()),
			zap.Error(err),
		)
		return err
	}

	logger.DebugContext(ctx, "user saved successfully",
		zap.String("user_id", string(u.ID())),
	)

	return nil
}

var _ user.UserRepository = (*UserRepository)(nil)
