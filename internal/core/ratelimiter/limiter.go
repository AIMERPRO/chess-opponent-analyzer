package ratelimiter

import (
	"sync"

	"golang.org/x/time/rate"
)

type IPRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

type GlobalRateLimiter struct {
	limiter *rate.Limiter
}

func NewGlobalRateLimiter(maxRate float64, burst int) *GlobalRateLimiter {
	return &GlobalRateLimiter{
		limiter: rate.NewLimiter(rate.Limit(maxRate), burst),
	}
}

func NewIPRateLimiter(maxRate float64, burst int) *IPRateLimiter {
	return &IPRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(maxRate),
		burst:    burst,
	}
}

func (g *GlobalRateLimiter) Allow() bool {
	return g.limiter.Allow()
}

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	limiter, exists := i.limiters[ip]
	i.mu.RUnlock()

	if !exists {
		i.mu.Lock()
		if _, exists = i.limiters[ip]; !exists {
			limiter = rate.NewLimiter(i.rate, i.burst)
			i.limiters[ip] = limiter
		}
		i.mu.Unlock()
	}

	return limiter
}

func (i *IPRateLimiter) Allow(ip string) bool {
	return i.GetLimiter(ip).Allow()
}

func (i *IPRateLimiter) Cleanup() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.limiters = make(map[string]*rate.Limiter)
}
