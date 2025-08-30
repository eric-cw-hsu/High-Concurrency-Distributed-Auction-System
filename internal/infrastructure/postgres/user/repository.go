package user

import (
	"context"
	"database/sql"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/user"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) GetUserById(ctx context.Context, userId string) (*user.User, error) {
	query := `SELECT id, email, password_hash, created_at FROM users WHERE id = $1`
	row := r.db.QueryRow(query, userId)

	var user user.User
	err := row.Scan(&user.Id, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, err // Other error
	}

	return &user, nil
}

func (r *PostgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `SELECT id, email, name, password_hash, created_at FROM users WHERE email = $1`
	row := r.db.QueryRow(query, email)

	var user user.User
	err := row.Scan(&user.Id, &user.Email, &user.Name, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, err // Other error
	}

	return &user, nil
}

func (r *PostgresUserRepository) GetAllUsers(ctx context.Context) ([]*user.User, error) {
	query := `SELECT id, email, name, password_hash, created_at FROM users`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*user.User
	for rows.Next() {
		var user user.User
		if err := rows.Scan(&user.Id, &user.Email, &user.Name, &user.PasswordHash, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *PostgresUserRepository) SaveUser(ctx context.Context, userData *user.User) (*user.User, error) {
	query := `
		INSERT INTO users (id, email, name, password_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, name, password_hash, created_at
	`

	row := r.db.QueryRowContext(ctx, query, userData.Id, userData.Email, userData.Name, userData.PasswordHash)

	var savedUser user.User
	if err := row.Scan(&savedUser.Id, &savedUser.Email, &savedUser.Name, &savedUser.PasswordHash, &savedUser.CreatedAt); err != nil {
		return nil, err
	}

	return &savedUser, nil
}

func (r *PostgresUserRepository) UpdateUser(ctx context.Context, userData *user.User) (*user.User, error) {
	query := `
		UPDATE users
		SET email = $2, password_hash = $3, created_at = NOW()
		WHERE id = $1
		RETURNING id, email, password_hash, created_at
	`

	row := r.db.QueryRowContext(ctx, query, userData.Id, userData.Email, userData.PasswordHash)

	var updatedUser user.User
	if err := row.Scan(&updatedUser.Id, &updatedUser.Email, &updatedUser.PasswordHash, &updatedUser.CreatedAt); err != nil {
		return nil, err
	}

	return &updatedUser, nil
}
