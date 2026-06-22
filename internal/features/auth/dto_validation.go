package auth

import (
	"fmt"

	"github.com/AIMERPRO/chess-opponent-analyzer/internal/core/apperrors"
)

// Validate
/*
	LoginRequestDTO.Validate() validates data from Login Request
	Username must be at least 5 characters and not null
	Password must be at least 6 characters and not null
	DeviceID must be not null
*/
func (req LoginRequestDTO) Validate() error {
	err := validateUsername(req.Username)
	if err != nil {
		return err
	}

	err = validatePassword(req.Password)
	if err != nil {
		return err
	}

	if req.DeviceID == "" {
		return fmt.Errorf("device_id is required: %w", apperrors.ErrInvalidInput)
	}

	return nil
}

// Validate
/*
	RegisterRequestDTO.Validate() validates data from Register Request
	Username must be at least 5 characters and not null
	Password must be at least 6 characters and not null
	LichessUsername must be not null
	DeviceID must be not null
*/
func (req RegisterRequestDTO) Validate() error {
	err := validateUsername(req.Username)
	if err != nil {
		return err
	}

	err = validatePassword(req.Password)
	if err != nil {
		return err
	}

	if req.LichessUsername == "" {
		return fmt.Errorf("lichess_username is required: %w", apperrors.ErrInvalidInput)
	}

	if req.DeviceID == "" {
		return fmt.Errorf("device_id is required: %w", apperrors.ErrInvalidInput)
	}
	return nil
}

// Validate
/*
	UpdateUserDTO.Validate() validates data from Update User Request
	Username must be at least 5 characters
*/
func (req UpdateUserDTO) Validate() error {
	if req.Username != nil {
		if len(*req.Username) < 5 {
			return fmt.Errorf("username must be at least 5 characters: %w", apperrors.ErrInvalidInput)
		}
	}

	return nil
}

func (req TokenRequestDTO) Validate() error {
	if req.RefreshToken == "" {
		return fmt.Errorf("refresh_token is required: %w", apperrors.ErrInvalidInput)
	}

	return nil
}

func validateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username is required: %w", apperrors.ErrInvalidInput)
	}

	if len(username) < 5 {
		return fmt.Errorf("username must be at least 5 characters: %w", apperrors.ErrInvalidInput)
	}

	return nil
}

func validatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password is required: %w", apperrors.ErrInvalidInput)
	}
	if len(password) < 6 {
		return fmt.Errorf("password must be at least 6 characters: %w", apperrors.ErrInvalidInput)
	}

	return nil
}
