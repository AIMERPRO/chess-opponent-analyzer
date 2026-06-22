package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/AIMERPRO/chess-opponent-analyzer/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository defines the interface for user data access operations
type UserRepository interface {
	// GetByID returns a user by their ID
	GetByID(ctx context.Context, id int64) (*domain.User, error)

	// GetByUsername returns a user by their Username
	GetByUsername(ctx context.Context, username string) (*domain.User, error)

	// Create creates a new user and returns the created user
	Create(ctx context.Context, user *domain.User) (*domain.User, error)

	// Update updates username and/or lichess username for a user
	Update(ctx context.Context, id int64, username *string, lichessUsername *string) (*domain.User, error)

	// Delete deletes a user by their ID
	Delete(ctx context.Context, id int64) error
}

type UserRepo struct {
	pool *pgxpool.Pool
}

// Just checks if UserRepo implements all methods of UserRepository interface
var _ UserRepository = (*UserRepo)(nil)

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, "SELECT id, username, lichess_username, password, created_at FROM users WHERE id = $1", id)

	var user domain.User
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.LichessUsername,
		&user.Password,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return &user, nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, "SELECT id, username, lichess_username, password, created_at FROM users WHERE username = $1", username)

	var user domain.User
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.LichessUsername,
		&user.Password,
		&user.CreatedAt,
		)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}

	return &user, nil
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	row := r.pool.QueryRow(ctx,
		"INSERT INTO users (username, lichess_username, password) VALUES ($1, $2, $3) RETURNING id, username, lichess_username, password, created_at",
		user.Username, user.LichessUsername, user.Password,
	)

	var created domain.User
	err := row.Scan(
		&created.ID,
		&created.Username,
		&created.LichessUsername,
		&created.Password,
		&created.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &created, nil
}

func (r *UserRepo) Update(ctx context.Context, id int64, username *string, lichessUsername *string) (*domain.User, error) {
	user, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if username != nil {
		user.Username = *username
	}
	if lichessUsername != nil {
		user.LichessUsername = *lichessUsername
	}

	_, err = r.pool.Exec(ctx,
		"UPDATE users SET username = $1, lichess_username = $2 WHERE id = $3",
		user.Username, user.LichessUsername, id,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM users WHERE id = $1", id)

	return err
}
