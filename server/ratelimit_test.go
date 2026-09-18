package main

import (
	"testing"
	"time"

	"golang.org/x/time/rate"
)

// a limiter on a clock that only moves when the test moves it
func newTestLimiter(limit rate.Limit, burst int) (*ipRateLimiter, *time.Time) {
	clock := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	limiter := newIPRateLimiter(limit, burst)
	limiter.now = func() time.Time { return clock }
	return limiter, &clock
}

func TestIPRateLimiterAllowsBurstThenWaits(t *testing.T) {
	limiter, clock := newTestLimiter(1, 2)
	for i := range 2 {
		if allowed, _ := limiter.take("1.1.1.1"); !allowed {
			t.Fatalf("expected request %d of the burst to be allowed", i+1)
		}
	}
	allowed, wait := limiter.take("1.1.1.1")
	if allowed || wait != time.Second {
		t.Fatalf("expected to be told to wait 1s after the burst, got allowed=%v wait=%v", allowed, wait)
	}

	*clock = clock.Add(time.Second)
	if allowed, _ := limiter.take("1.1.1.1"); !allowed {
		t.Error("expected a token to be back after waiting")
	}
}

func TestIPRateLimiterRejectionsDontUseUpTokens(t *testing.T) {
	limiter, _ := newTestLimiter(1, 1)
	limiter.take("1.1.1.1")
	// if rejected requests still queued for a token, each one would push the wait further out
	for range 5 {
		if allowed, wait := limiter.take("1.1.1.1"); allowed || wait != time.Second {
			t.Fatalf("expected every rejection to wait the same 1s, got allowed=%v wait=%v", allowed, wait)
		}
	}
}

func TestIPRateLimiterTracksIPsSeparately(t *testing.T) {
	limiter, _ := newTestLimiter(1, 1)
	limiter.take("1.1.1.1")
	if allowed, _ := limiter.take("1.1.1.1"); allowed {
		t.Fatal("expected first ip to be out of tokens")
	}
	if allowed, _ := limiter.take("2.2.2.2"); !allowed {
		t.Error("expected a different ip to have its own bucket")
	}
}

func TestIPRateLimiterForgetsQuietIPs(t *testing.T) {
	limiter, clock := newTestLimiter(1, 1)
	limiter.take("1.1.1.1")
	*clock = clock.Add(visitorTTL + time.Second)
	limiter.take("2.2.2.2")
	if _, remembered := limiter.visitors["1.1.1.1"]; remembered || len(limiter.visitors) != 1 {
		t.Errorf("expected only the recent ip to be kept, got %v", limiter.visitors)
	}
}
