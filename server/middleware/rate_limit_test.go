package middleware

import (
	"context"
	"testing"
	"time"
)

func TestMemoryRateLimiter_FirstRequestAllowed(t *testing.T) {
	rl := NewMemoryRateLimiter()
	if !rl.Allow(context.Background(), "user1", 5, time.Minute) {
		t.Error("first request should be allowed")
	}
}

func TestMemoryRateLimiter_WithinLimit(t *testing.T) {
	rl := NewMemoryRateLimiter()
	for i := 0; i < 5; i++ {
		if !rl.Allow(context.Background(), "user1", 5, time.Minute) {
			t.Errorf("request %d within limit should be allowed", i+1)
		}
	}
}

func TestMemoryRateLimiter_ExceedsLimit(t *testing.T) {
	rl := NewMemoryRateLimiter()
	for i := 0; i < 5; i++ {
		rl.Allow(context.Background(), "user1", 5, time.Minute)
	}
	if rl.Allow(context.Background(), "user1", 5, time.Minute) {
		t.Error("6th request should be denied")
	}
}

func TestMemoryRateLimiter_DifferentKeysIndependent(t *testing.T) {
	rl := NewMemoryRateLimiter()
	// drain user1
	for i := 0; i < 5; i++ {
		rl.Allow(context.Background(), "user1", 5, time.Minute)
	}
	// user2 should still be allowed
	if !rl.Allow(context.Background(), "user2", 5, time.Minute) {
		t.Error("user2 should be independent of user1's limit")
	}
	// user1 should be denied
	if rl.Allow(context.Background(), "user1", 5, time.Minute) {
		t.Error("user1 should still be over limit")
	}
}

func TestMemoryRateLimiter_WindowReset(t *testing.T) {
	rl := NewMemoryRateLimiter()
	// Drain the limit
	for i := 0; i < 5; i++ {
		rl.Allow(context.Background(), "user1", 5, 10*time.Millisecond)
	}
	if rl.Allow(context.Background(), "user1", 5, 10*time.Millisecond) {
		t.Error("should be denied before window expires")
	}
	// Wait for the window to expire
	time.Sleep(15 * time.Millisecond)
	if !rl.Allow(context.Background(), "user1", 5, 10*time.Millisecond) {
		t.Error("should be allowed after window expires")
	}
}

func TestMemoryRateLimiter_ZeroLimit(t *testing.T) {
	rl := NewMemoryRateLimiter()
	// limit=0: first request allowed, second denied
	if !rl.Allow(context.Background(), "user1", 0, time.Minute) {
		t.Error("first request with limit=0 should be allowed")
	}
	if rl.Allow(context.Background(), "user1", 0, time.Minute) {
		t.Error("second request with limit=0 should be denied (count=1 >= limit=0)")
	}
}

func TestRateLimit_InterfaceContract(t *testing.T) {
	var rl RateLimiter
	rl = NewMemoryRateLimiter()
	if rl == nil {
		t.Error("NewMemoryRateLimiter returned nil")
	}
}
