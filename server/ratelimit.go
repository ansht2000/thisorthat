package main

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// how long an ip can go quiet before its bucket is dropped. well past the time
// a bucket takes to refill completely, so forgetting one changes nothing
const visitorTTL = 10 * time.Minute

// gives each client ip its own token bucket
type ipRateLimiter struct {
	mu        sync.Mutex
	visitors  map[string]*visitor
	limit     rate.Limit
	burst     int
	lastSweep time.Time
	// swappable so tests can move time forward without sleeping
	now func() time.Time
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// each ip can make burst requests at once, then limit more per second after that
func newIPRateLimiter(limit rate.Limit, burst int) *ipRateLimiter {
	return &ipRateLimiter{
		visitors: map[string]*visitor{},
		limit:    limit,
		burst:    burst,
		now:      time.Now,
	}
}

// takes a token from ip's bucket. if the bucket is empty it returns
// false and how long until the next token is available
func (l *ipRateLimiter) take(ip string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	// sweeping inline instead of on a background goroutine keeps the map from
	// growing forever without leaving anything running that would need stopping
	if now.Sub(l.lastSweep) > visitorTTL {
		for key, v := range l.visitors {
			if now.Sub(v.lastSeen) > visitorTTL {
				delete(l.visitors, key)
			}
		}
		l.lastSweep = now
	}

	v, ok := l.visitors[ip]
	if !ok {
		v = &visitor{limiter: rate.NewLimiter(l.limit, l.burst)}
		l.visitors[ip] = v
	}
	v.lastSeen = now

	reservation := v.limiter.ReserveN(now, 1)
	if wait := reservation.DelayFrom(now); wait > 0 {
		// hand the token back, this request isn't going to wait for it
		reservation.CancelAt(now)
		return false, wait
	}
	return true, 0
}
