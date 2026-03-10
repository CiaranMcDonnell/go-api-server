package middleware

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/ciaranmcdonnell/go-api-server/internal/metrics"
	"github.com/ciaranmcdonnell/go-api-server/pkg/apperrors"
	"github.com/gin-gonic/gin"
)

type limiterEntry struct {
	tokens    float64
	lastCheck time.Time
}

type RateLimiterConfig struct {
	Rate       float64       // tokens per second (e.g. 0.2 = 1 request per 5 seconds)
	Burst      int           // max burst size
	CleanupTTL time.Duration // how long to keep idle entries
}

type rateLimiter struct {
	mu      sync.Mutex
	entries map[string]*limiterEntry
	config  RateLimiterConfig
	stop    chan struct{}
}

func newRateLimiter(config RateLimiterConfig) *rateLimiter {
	rl := &rateLimiter{
		entries: make(map[string]*limiterEntry),
		config:  config,
		stop:    make(chan struct{}),
	}

	go func() {
		ticker := time.NewTicker(config.CleanupTTL)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				rl.evict()
			case <-rl.stop:
				return
			}
		}
	}()

	return rl
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	e, exists := rl.entries[key]
	if !exists {
		rl.entries[key] = &limiterEntry{
			tokens:    float64(rl.config.Burst) - 1,
			lastCheck: now,
		}
		return true
	}

	elapsed := now.Sub(e.lastCheck).Seconds()
	e.tokens += elapsed * rl.config.Rate
	if e.tokens > float64(rl.config.Burst) {
		e.tokens = float64(rl.config.Burst)
	}
	e.lastCheck = now

	if e.tokens >= 1 {
		e.tokens--
		return true
	}
	return false
}

func (rl *rateLimiter) evict() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-rl.config.CleanupTTL)
	for key, e := range rl.entries {
		if e.lastCheck.Before(cutoff) {
			delete(rl.entries, key)
		}
	}
}

// RateLimit returns a Gin middleware that rate-limits by client IP.
// Typical auth usage: Rate=0.2 (1 req/5s sustained), Burst=5 (short bursts OK).
func RateLimit(config RateLimiterConfig) gin.HandlerFunc {
	rl := newRateLimiter(config)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !rl.allow(ip) {
			metrics.RateLimitRejectsTotal.WithLabelValues(c.FullPath()).Inc()
			slog.Warn("rate limit exceeded",
				"ip", ip,
				"path", c.FullPath(),
			)
			apperrors.Error(c, http.StatusTooManyRequests, "rate_limited", "Too many requests, please try again later")
			c.Abort()
			return
		}

		c.Next()
	}
}
