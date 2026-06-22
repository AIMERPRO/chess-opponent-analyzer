package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AIMERPRO/chess-opponent-analyzer/internal/core/config"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret"

// signToken builds an HS256 token signed with the given secret.
func signToken(t *testing.T, secret string, userID int64, exp time.Time) string {
	t.Helper()
	claims := UserIDClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("signing test token failed: %v", err)
	}
	return signed
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	cfg := &config.Config{JwtSecret: testSecret}

	var (
		nextCalled bool
		gotUserID  int64
	)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		id, ok := r.Context().Value(UserIDKey).(int64)
		if !ok {
			t.Error("user id missing from request context")
		}
		gotUserID = id
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/analyze/foo", nil)
	req.Header.Set("Authorization", "Bearer "+signToken(t, testSecret, 42, time.Now().Add(time.Hour)))
	rr := httptest.NewRecorder()

	AuthMiddleware(cfg, next).ServeHTTP(rr, req)

	if !nextCalled {
		t.Error("next handler was not called for a valid token")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	if gotUserID != 42 {
		t.Errorf("user id in context = %d, want 42", gotUserID)
	}
}

func TestAuthMiddleware_Unauthorized(t *testing.T) {
	cfg := &config.Config{JwtSecret: testSecret}

	tests := []struct {
		name       string
		authHeader string
	}{
		{name: "missing header", authHeader: ""},
		{name: "malformed token", authHeader: "Bearer not-a-jwt"},
		{name: "wrong secret", authHeader: "Bearer " + signToken(t, "another-secret", 42, time.Now().Add(time.Hour))},
		{name: "expired token", authHeader: "Bearer " + signToken(t, testSecret, 42, time.Now().Add(-time.Hour))},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				nextCalled = true
			})

			req := httptest.NewRequest(http.MethodGet, "/analyze/foo", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rr := httptest.NewRecorder()

			AuthMiddleware(cfg, next).ServeHTTP(rr, req)

			if nextCalled {
				t.Error("next handler should not be called for an invalid token")
			}
			if rr.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want 401", rr.Code)
			}
		})
	}
}
