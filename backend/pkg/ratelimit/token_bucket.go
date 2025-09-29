package ratelimit

import (
	"math"
	"sync"
	"time"
)

type TokenBucket struct {
	mu         sync.Mutex
	capacity   float64
	tokens     float64
	refillRate float64
	lastRefill time.Time
}

func NewTokenBucket(perMinute int) *TokenBucket {
	if perMinute <= 0 {
		perMinute = 1
	}
	rate := float64(perMinute) / 60.0
	now := time.Now()
	return &TokenBucket{
		capacity:   float64(perMinute),
		tokens:     float64(perMinute),
		refillRate: rate,
		lastRefill: now,
	}
}

func (tb *TokenBucket) refill(now time.Time) {
	elapsed := now.Sub(tb.lastRefill).Seconds()
	if elapsed <= 0 {
		return
	}
	tb.tokens = math.Min(tb.capacity, tb.tokens+elapsed*tb.refillRate)
	tb.lastRefill = now
}

func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill(time.Now())
	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}
	return false
}

func (tb *TokenBucket) Remaining() int {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill(time.Now())
	return int(math.Floor(tb.tokens + 1e-9))
}
