package app

import (
	"context"
	"time"

	"github.com/AIMERPRO/chess-opponent-analyzer/internal/core/ratelimiter"
	"github.com/AIMERPRO/chess-opponent-analyzer/internal/features/auth"
	"go.uber.org/zap"
)

func startExpiredTokensCleaner(ctx context.Context, tokenRepo auth.TokenRepository, log *zap.Logger) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := tokenRepo.DeleteExpiredTokens(ctx); err != nil {
				log.Error("failed to delete expired tokens", zap.Error(err))
			}
		case <-ctx.Done():
			return
		}
	}
}

func startIPLimiterCleaner(ctx context.Context, ipRateLimiter *ratelimiter.IPRateLimiter, log *zap.Logger) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ipRateLimiter.Cleanup()
		case <-ctx.Done():
			return
		}
	}
}
