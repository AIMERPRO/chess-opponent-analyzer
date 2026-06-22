package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AIMERPRO/chess-opponent-analyzer/internal/core/config"
	"github.com/AIMERPRO/chess-opponent-analyzer/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// mockTokenRepo is a hand-written TokenRepository mock for helper tests.
// Only Create is exercised by generateTokenPair; the rest satisfy the interface.
type mockTokenRepo struct {
	created   *domain.RefreshToken
	createErr error
}

func (m *mockTokenRepo) Create(_ context.Context, token *domain.RefreshToken) (*domain.RefreshToken, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	token.ID = 1
	m.created = token
	return token, nil
}

func (m *mockTokenRepo) GetTokenByValue(context.Context, string) (*domain.RefreshToken, error) {
	return nil, errors.New("not implemented")
}
func (m *mockTokenRepo) Delete(context.Context, int64) error              { return nil }
func (m *mockTokenRepo) DeleteAllUserTokens(context.Context, int64) error { return nil }
func (m *mockTokenRepo) DeleteExpiredTokens(context.Context) error        { return nil }

const testJwtSecret = "test-secret"

func newTestService(tokenRepo TokenRepository) *service {
	return &service{
		tokenRepo: tokenRepo,
		cfg:       &config.Config{JwtSecret: testJwtSecret},
	}
}

func TestService_HashAndComparePassword(t *testing.T) {
	s := &service{}

	hash, err := s.hashPassword("secret123")
	if err != nil {
		t.Fatalf("hashPassword() error = %v", err)
	}
	if hash == "" {
		t.Fatal("hashPassword() returned empty hash")
	}
	if hash == "secret123" {
		t.Fatal("hashPassword() returned the plaintext password")
	}

	ok, err := s.comparePassword("secret123", hash)
	if err != nil {
		t.Fatalf("comparePassword() error = %v", err)
	}
	if !ok {
		t.Fatal("comparePassword() = false for the correct password")
	}

	ok, _ = s.comparePassword("wrong-password", hash)
	if ok {
		t.Fatal("comparePassword() = true for a wrong password")
	}
}

func TestService_GenerateAccessToken(t *testing.T) {
	s := newTestService(&mockTokenRepo{})

	tokenString, err := s.generateAccessToken(42)
	if err != nil {
		t.Fatalf("generateAccessToken() error = %v", err)
	}

	parsed, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(testJwtSecret), nil
	})
	if err != nil {
		t.Fatalf("parsing generated token failed: %v", err)
	}
	if !parsed.Valid {
		t.Fatal("generated token is not valid")
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("unexpected claims type %T", parsed.Claims)
	}
	if claims["user_id"] != float64(42) {
		t.Errorf("user_id claim = %v, want 42", claims["user_id"])
	}
}

func TestService_GenerateAccessToken_WrongSecretFailsVerification(t *testing.T) {
	s := newTestService(&mockTokenRepo{})

	tokenString, err := s.generateAccessToken(1)
	if err != nil {
		t.Fatalf("generateAccessToken() error = %v", err)
	}

	_, err = jwt.Parse(tokenString, func(*jwt.Token) (interface{}, error) {
		return []byte("a-different-secret"), nil
	})
	if err == nil {
		t.Fatal("expected verification error with a wrong secret, got nil")
	}
}

func TestService_GenerateRefreshToken(t *testing.T) {
	s := &service{}

	first, err := s.generateRefreshToken()
	if err != nil {
		t.Fatalf("generateRefreshToken() error = %v", err)
	}
	if _, err = uuid.Parse(first); err != nil {
		t.Errorf("generateRefreshToken() returned a non-UUID value %q: %v", first, err)
	}

	second, err := s.generateRefreshToken()
	if err != nil {
		t.Fatalf("generateRefreshToken() error = %v", err)
	}
	if first == second {
		t.Error("generateRefreshToken() returned the same value twice")
	}
}

func TestService_GenerateTokenPair(t *testing.T) {
	repo := &mockTokenRepo{}
	s := newTestService(repo)

	pair, err := s.generateTokenPair(context.Background(), 7, "device-1")
	if err != nil {
		t.Fatalf("generateTokenPair() error = %v", err)
	}

	if pair.AccessToken == "" {
		t.Error("generateTokenPair() returned empty access token")
	}
	if pair.RefreshToken == "" {
		t.Error("generateTokenPair() returned empty refresh token")
	}
	if repo.created == nil {
		t.Fatal("tokenRepo.Create was not called")
	}
	if pair.RefreshToken != repo.created.Token {
		t.Errorf("pair.RefreshToken = %q, want stored token %q", pair.RefreshToken, repo.created.Token)
	}
	if repo.created.UserID != 7 || repo.created.DeviceID != "device-1" {
		t.Errorf("stored token has userID=%d deviceID=%q, want 7/device-1", repo.created.UserID, repo.created.DeviceID)
	}

	// refresh token should live ~30 days from now
	wantExp := time.Now().Add(30 * 24 * time.Hour)
	if diff := pair.ExpiresAt.Sub(wantExp); diff > time.Minute || diff < -time.Minute {
		t.Errorf("ExpiresAt = %v, want ~%v", pair.ExpiresAt, wantExp)
	}
}

func TestService_GenerateTokenPair_RepoError(t *testing.T) {
	repo := &mockTokenRepo{createErr: errors.New("db down")}
	s := newTestService(repo)

	_, err := s.generateTokenPair(context.Background(), 7, "device-1")
	if err == nil {
		t.Fatal("expected error when tokenRepo.Create fails, got nil")
	}
}
