package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/AIMERPRO/chess-opponent-analyzer/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TokenRepository defines the interface for refresh token data access
type TokenRepository interface {
	// GetTokenByValue returns a refresh token by value (refresh token)
	GetTokenByValue(ctx context.Context, value string) (*domain.RefreshToken, error)

	// Create creates a new token and returns the created refresh token
	Create(ctx context.Context, token *domain.RefreshToken) (*domain.RefreshToken, error)

	// Delete deletes a token by taken id
	Delete(ctx context.Context, id int64) error

	// DeleteAllUserTokens deletes all exists tokens from user (logout operation)
	DeleteAllUserTokens(ctx context.Context, userID int64) error
}

type TokenRepo struct {
	pool *pgxpool.Pool
}

var _ TokenRepository = (*TokenRepo)(nil)

func NewTokenRepo(pool *pgxpool.Pool) *TokenRepo {
	return &TokenRepo{
		pool: pool,
	}
}

func (r *TokenRepo) GetTokenByValue(ctx context.Context, value string) (*domain.RefreshToken, error) {
	row := r.pool.QueryRow(ctx, "SELECT id, user_id, device_id, token, expires_at, created_at FROM refresh_tokens WHERE token = $1", value)

	var token domain.RefreshToken

	err := row.Scan(
		&token.ID,
		&token.UserID,
		&token.DeviceID,
		&token.Token,
		&token.ExpiresAt,
		&token.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("token not found: %w", err)
		}
		return nil, fmt.Errorf("get token by value: %w", err)
	}

	return &token, nil
}

func (r *TokenRepo) Create(ctx context.Context, token *domain.RefreshToken) (*domain.RefreshToken, error) {
	query := `INSERT INTO refresh_tokens (user_id, device_id, token, expires_at)
			  VALUES ($1, $2, $3, $4)
			  ON CONFLICT (user_id, device_id) 
			  DO UPDATE SET token = $3, expires_at = $4
			  RETURNING id, user_id, device_id, token, expires_at, created_at`

	row := r.pool.QueryRow(ctx, query, token.UserID, token.DeviceID, token.Token, token.ExpiresAt)

	var created domain.RefreshToken
	err := row.Scan(
		&created.ID,
		&created.UserID,
		&created.DeviceID,
		&created.Token,
		&created.ExpiresAt,
		&created.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &created, err
}

func (r *TokenRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM refresh_tokens WHERE id = $1", id)

	return err
}

func (r *TokenRepo) DeleteAllUserTokens(ctx context.Context, userID int64) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM refresh_tokens WHERE user_id = $1", userID)

	return err
}
