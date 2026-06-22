package auth

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"time"

	"github.com/AIMERPRO/chess-opponent-analyzer/internal/core/response"
	"github.com/AIMERPRO/chess-opponent-analyzer/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func (s *service) hashPassword(password string) (string, error) {
	hash := sha256.Sum256([]byte(password))
	hashedPassword, err := bcrypt.GenerateFromPassword(hash[:], bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

func (s *service) comparePassword(password string, hash string) (bool, error) {
	sha256Hash := sha256.Sum256([]byte(password))
	err := bcrypt.CompareHashAndPassword([]byte(hash), sha256Hash[:])

	return err == nil, err
}

func (s *service) generateAccessToken(userID int64) (string, error) {
	type UserIDClaims struct {
		UserID int64 `json:"user_id"`
		jwt.RegisteredClaims
	}

	claims := UserIDClaims{
		userID,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(s.cfg.JwtSecret))

	return tokenString, err
}

func (s *service) generateRefreshToken() (string, error) {
	tokenUUID, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return tokenUUID.String(), nil
}

func (s *service) generateTokenPair(ctx context.Context, userID int64, deviceID string) (TokenResponseDTO, error) {
	accessToken, err := s.generateAccessToken(userID)
	if err != nil {
		return TokenResponseDTO{}, err
	}

	refreshToken, err := s.generateRefreshToken()
	if err != nil {
		return TokenResponseDTO{}, err
	}

	refreshTokenModel := domain.RefreshToken{
		UserID:    userID,
		Token:     refreshToken,
		DeviceID:  deviceID,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}

	createdToken, err := s.tokenRepo.Create(ctx, &refreshTokenModel)
	if err != nil {
		return TokenResponseDTO{}, err
	}

	tokenPair := TokenResponseDTO{
		AccessToken:  accessToken,
		RefreshToken: createdToken.Token,
		ExpiresAt:    createdToken.ExpiresAt,
	}

	return tokenPair, nil
}

func decodeAndValidate[T validatable](r *http.Request, w http.ResponseWriter, log *zap.Logger) (T, bool) {
	var req T
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("decoding request failed", zap.Error(err))
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return req, false
	}

	if err := req.Validate(); err != nil {
		log.Error("validation failed", zap.Error(err))
		response.Error(w, http.StatusBadRequest, err.Error())
		return req, false
	}

	return req, true
}
