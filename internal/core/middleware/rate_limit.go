package middleware

import (
	"net"
	"net/http"

	"github.com/AIMERPRO/chess-opponent-analyzer/internal/core/ratelimiter"
)

func RateLimitMiddleware(global *ratelimiter.GlobalRateLimiter, ip *ratelimiter.IPRateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// вот тут r уже есть — это конкретный входящий запрос
		clientIP := getIP(r)

		if !global.Allow() {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}

		if !ip.Allow(clientIP) {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	return ip
}
