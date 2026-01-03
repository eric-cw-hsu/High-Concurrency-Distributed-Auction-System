package postgres

import (
	"database/sql"
	"errors"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/domain/user"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/kernel"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) user.UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByEmail(email string) (*user.User, error) {
	var id, passwordHash, status string

	row := r.db.QueryRow("SELECT id, password_hash, status FROM users WHERE email = $1", email)
	if err := row.Scan(&id, &passwordHash, &status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	userID, err := kernel.NewUserID(id)
	if err != nil {
		return nil, err
	}

	user := user.NewUser(
		userID,
		email,
		passwordHash,
		user.UserStatus(status),
	)

	return user, nil
}

func (r *UserRepository) Save(user *user.User) error {
	_, err := r.db.Exec(
		"INSERT INTO users (id, email, password_hash, status) VALUES ($1, $2, $3, $4)",
		user.ID(),
		user.Email(),
		user.PasswordHash(),
		user.Status(),
	)
	return err
}

var _ user.UserRepository = (*UserRepository)(nil)
