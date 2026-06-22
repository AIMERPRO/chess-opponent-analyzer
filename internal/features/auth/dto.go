package auth

import "time"

// LoginRequestDTO represents the request body for user authentication
type LoginRequestDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
	DeviceID string `json:"device_id"`
}

// RegisterRequestDTO represents the request body for user registration
type RegisterRequestDTO struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	LichessUsername string `json:"lichess_username"`
	DeviceID        string `json:"device_id"`
}

// UpdateUserDTO represents the request body for update user username or lichess username
type UpdateUserDTO struct {
	Username        *string `json:"username"`
	LichessUsername *string `json:"lichess_username"`
}

// TokenRequestDTO represents the request body containing a refresh token
type TokenRequestDTO struct {
	RefreshToken string `json:"refresh_token"`
}

// TokenResponseDTO represents the response body for client with new refresh and access token
type TokenResponseDTO struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// UserResponseDTO represents the response body for client with user data
type UserResponseDTO struct {
	ID              int       `json:"id"`
	Username        string    `json:"username"`
	LichessUsername string    `json:"lichess_username"`
	CreatedAt       time.Time `json:"created_at"`
}
