package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/AIMERPRO/chess-opponent-analyzer/internal/core/config"
	"github.com/AIMERPRO/chess-opponent-analyzer/internal/domain"
)

// Service defines the interface for auth business logic
type Service interface {
	// Register creates a new user account and returns authentication tokens
	Register(ctx context.Context, req RegisterRequestDTO) (TokenResponseDTO, error)

	// Login authenticates a user by credentials and returns authentication tokens
	Login(ctx context.Context, req LoginRequestDTO) (TokenResponseDTO, error)

	// RefreshToken issues a new access token using a valid refresh token
	RefreshToken(ctx context.Context, req TokenRequestDTO) (TokenResponseDTO, error)

	// GetUser returns public information about a user by their ID
	GetUser(ctx context.Context, userID int64) (UserResponseDTO, error)

	// UpdateUser updates user profile fields and returns updated user info
	UpdateUser(ctx context.Context, userID int64, req UpdateUserDTO) (UserResponseDTO, error)

	// Logout invalidates the provided refresh token
	Logout(ctx context.Context, refreshToken string) error

	// LogoutFromAllDevices invalidates all refresh tokens for the given user
	LogoutFromAllDevices(ctx context.Context, userID int64) error
}

type service struct {
	userRepo  UserRepository
	tokenRepo TokenRepository
	cfg       *config.Config
}

var _ Service = (*service)(nil)

func NewService(userRepo UserRepository, tokenRepo TokenRepository, cfg *config.Config) Service {
	return &service{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		cfg:       cfg,
	}
}

func (s *service) Register(ctx context.Context, req RegisterRequestDTO) (TokenResponseDTO, error) {
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return TokenResponseDTO{}, err
	}

	userToCreate := domain.User{
		Username:        req.Username,
		Password:        hashedPassword,
		LichessUsername: req.LichessUsername,
	}

	user, err := s.userRepo.Create(ctx, &userToCreate)
	if err != nil {
		return TokenResponseDTO{}, err
	}

	tokenPair, err := s.generateTokenPair(ctx, user.ID, req.DeviceID)
	if err != nil {
		return TokenResponseDTO{}, err
	}

	return tokenPair, nil
}

func (s *service) Login(ctx context.Context, req LoginRequestDTO) (TokenResponseDTO, error) {
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return TokenResponseDTO{}, err
	}

	passwordValidate, err := s.comparePassword(req.Password, user.Password)
	if err != nil {
		return TokenResponseDTO{}, err
	}

	if !passwordValidate {
		return TokenResponseDTO{}, fmt.Errorf("invalid credentials")
	}

	tokenPair, err := s.generateTokenPair(ctx, user.ID, req.DeviceID)
	if err != nil {
		return TokenResponseDTO{}, err
	}

	return tokenPair, nil
}

func (s *service) RefreshToken(ctx context.Context, req TokenRequestDTO) (TokenResponseDTO, error) {
	token, err := s.tokenRepo.GetTokenByValue(ctx, req.RefreshToken)
	if err != nil {
		return TokenResponseDTO{}, err
	}

	if time.Now().After(token.ExpiresAt) {
		err = s.tokenRepo.Delete(ctx, token.ID)
		if err != nil {
			return TokenResponseDTO{}, fmt.Errorf("failed to delete expired token: %w", err)
		}
		return TokenResponseDTO{}, fmt.Errorf("refresh token expired, please login again")
	}

	tokenPair, err := s.generateTokenPair(ctx, token.UserID, token.DeviceID)
	if err != nil {
		return TokenResponseDTO{}, err
	}

	return tokenPair, nil
}

func (s *service) GetUser(ctx context.Context, userID int64) (UserResponseDTO, error) {
	user, err := s.userRepo.GetByID(ctx, userID)

	if err != nil {
		return UserResponseDTO{}, err
	}

	return UserResponseDTO{
		ID:              int(user.ID),
		Username:        user.Username,
		LichessUsername: user.LichessUsername,
		CreatedAt:       user.CreatedAt,
	}, nil
}

func (s *service) UpdateUser(ctx context.Context, userID int64, req UpdateUserDTO) (UserResponseDTO, error) {
	user, err := s.userRepo.Update(ctx, userID, req.Username, req.LichessUsername)

	if err != nil {
		return UserResponseDTO{}, err
	}

	return UserResponseDTO{
		ID:              int(user.ID),
		Username:        user.Username,
		LichessUsername: user.LichessUsername,
		CreatedAt:       user.CreatedAt,
	}, nil
}

func (s *service) Logout(ctx context.Context, refreshToken string) error {
	token, err := s.tokenRepo.GetTokenByValue(ctx, refreshToken)
	if err != nil {
		return err
	}

	err = s.tokenRepo.Delete(ctx, token.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) LogoutFromAllDevices(ctx context.Context, userID int64) error {
	err := s.tokenRepo.DeleteAllUserTokens(ctx, userID)
	if err != nil {
		return err
	}

	return nil
}
