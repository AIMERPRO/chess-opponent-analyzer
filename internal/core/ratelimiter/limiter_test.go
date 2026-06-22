package ratelimiter

import "testing"

func TestGlobalRateLimiter_Allow(t *testing.T) {
	// rate 0 => no refill, so only the initial burst of tokens is allowed.
	g := NewGlobalRateLimiter(0, 2)

	if !g.Allow() {
		t.Error("1st Allow() = false, want true")
	}
	if !g.Allow() {
		t.Error("2nd Allow() = false, want true")
	}
	if g.Allow() {
		t.Error("3rd Allow() = true, want false (burst exhausted)")
	}
}

func TestIPRateLimiter_PerIP(t *testing.T) {
	i := NewIPRateLimiter(0, 1)

	if !i.Allow("1.1.1.1") {
		t.Error("first request from IP a = false, want true")
	}
	if i.Allow("1.1.1.1") {
		t.Error("second request from IP a = true, want false (burst exhausted)")
	}
	// a different IP has its own independent bucket
	if !i.Allow("2.2.2.2") {
		t.Error("first request from IP b = false, want true")
	}
}

func TestIPRateLimiter_GetLimiterReuse(t *testing.T) {
	i := NewIPRateLimiter(1, 1)

	first := i.GetLimiter("1.1.1.1")
	again := i.GetLimiter("1.1.1.1")
	if first != again {
		t.Error("GetLimiter() returned a different limiter for the same IP")
	}

	other := i.GetLimiter("2.2.2.2")
	if first == other {
		t.Error("GetLimiter() returned the same limiter for different IPs")
	}
}

func TestIPRateLimiter_Cleanup(t *testing.T) {
	i := NewIPRateLimiter(0, 1)

	if !i.Allow("1.1.1.1") {
		t.Fatal("setup: first Allow() = false, want true")
	}
	if i.Allow("1.1.1.1") {
		t.Fatal("setup: bucket should be exhausted")
	}

	i.Cleanup()

	if !i.Allow("1.1.1.1") {
		t.Error("after Cleanup() Allow() = false, want true (bucket reset)")
	}
}
