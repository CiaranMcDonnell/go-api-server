package middleware

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/ciaranmcdonnell/go-api-server/internal/metrics"
	"github.com/ciaranmcdonnell/go-api-server/pkg/apperrors"
	"github.com/ciaranmcdonnell/go-api-server/pkg/utils"
	"github.com/gin-gonic/gin"
)

// --- Key functions ---

// KeyByIP keys rate limiting on the client IP address.
// Use for unauthenticated endpoints (login, register).
func KeyByIP(c *gin.Context) string {
	return "ip:" + c.ClientIP()
}

// KeyByUserID keys rate limiting on the authenticated user ID from context.
// Falls back to IP if no user ID is present (e.g. auth failed silently).
func KeyByUserID(c *gin.Context) string {
	if uid, ok := c.Get(utils.ContextKeyUserID); ok {
		if s, ok := uid.(string); ok && s != "" {
			return "user:" + s
		}
	}
	return "ip:" + c.ClientIP()
}

// --- Config ---

// RateLimiterConfig configures a rate limiter instance.
type RateLimiterConfig struct {
	Rate            float64                   // tokens per second
	Burst           int                       // max burst size
	CleanupInterval time.Duration             // how often the eviction goroutine runs
	IdleTTL         time.Duration             // how long an entry can sit idle before eviction
	KeyFunc         func(*gin.Context) string // determines the bucket key per request
	TierName        string                    // "strict", "standard", "relaxed" — for metrics
}

func (c *RateLimiterConfig) applyDefaults() {
	if c.CleanupInterval == 0 {
		c.CleanupInterval = 2 * time.Minute
	}
	if c.IdleTTL == 0 {
		c.IdleTTL = 10 * time.Minute
	}
	if c.IdleTTL < c.CleanupInterval {
		slog.Warn("rate limiter: IdleTTL < CleanupInterval, adjusting IdleTTL to match",
			"idle_ttl", c.IdleTTL,
			"cleanup_interval", c.CleanupInterval,
		)
		c.IdleTTL = c.CleanupInterval
	}
	if c.KeyFunc == nil {
		c.KeyFunc = KeyByIP
	}
}

// --- Token bucket ---

type limiterEntry struct {
	tokens    float64
	lastCheck time.Time
}

type rateLimiter struct {
	mu      sync.Mutex
	entries map[string]*limiterEntry
	config  RateLimiterConfig
	stop    chan struct{}
}

func newRateLimiter(config RateLimiterConfig) *rateLimiter {
	config.applyDefaults()

	rl := &rateLimiter{
		entries: make(map[string]*limiterEntry),
		config:  config,
		stop:    make(chan struct{}),
	}

	go func() {
		ticker := time.NewTicker(config.CleanupInterval)
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

	cutoff := time.Now().Add(-rl.config.IdleTTL)
	for key, e := range rl.entries {
		if e.lastCheck.Before(cutoff) {
			delete(rl.entries, key)
		}
	}
}

// RateLimit returns a Gin middleware that rate-limits requests using a token bucket.
// The bucket key is determined by config.KeyFunc (IP, user ID, etc.).
func RateLimit(config RateLimiterConfig) gin.HandlerFunc {
	rl := newRateLimiter(config)

	return func(c *gin.Context) {
		key := rl.config.KeyFunc(c)

		if !rl.allow(key) {
			metrics.RateLimitRejectsTotal.WithLabelValues(c.FullPath(), rl.config.TierName).Inc()
			slog.Warn("rate limit exceeded",
				"key", key,
				"path", c.FullPath(),
				"tier", rl.config.TierName,
			)
			apperrors.Error(c, http.StatusTooManyRequests, "rate_limited", "Too many requests, please try again later")
			c.Abort()
			return
		}

		c.Next()
	}
}
