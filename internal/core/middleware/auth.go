package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/AIMERPRO/chess-opponent-analyzer/internal/core/config"
	"github.com/golang-jwt/jwt/v5"
)

type UserIDClaims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

type contextKey string

const UserIDKey contextKey = "user_id"

func AuthMiddleware(cfg *config.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwtToken := r.Header.Get("Authorization")
		jwtToken = strings.TrimPrefix(jwtToken, "Bearer ")

		var claims UserIDClaims
		_, err := jwt.ParseWithClaims(jwtToken, &claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.JwtSecret), nil
		})
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
