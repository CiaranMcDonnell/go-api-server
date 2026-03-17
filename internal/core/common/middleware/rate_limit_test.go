package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ciaranmcdonnell/go-api-server/internal/metrics"
	"github.com/ciaranmcdonnell/go-api-server/pkg/utils"
	"github.com/gin-gonic/gin"
	dto "github.com/prometheus/client_model/go"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// --- Key function tests ---

func TestKeyByIP(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.RemoteAddr = "192.168.1.1:1234"

	key := KeyByIP(c)
	if key != "ip:192.168.1.1" {
		t.Errorf("KeyByIP = %q, want %q", key, "ip:192.168.1.1")
	}
}

func TestKeyByUserID_WithUser(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.RemoteAddr = "192.168.1.1:1234"
	c.Set(utils.ContextKeyUserID, "42")

	key := KeyByUserID(c)
	if key != "user:42" {
		t.Errorf("KeyByUserID = %q, want %q", key, "user:42")
	}
}

func TestKeyByUserID_NoUser_FallsBackToIP(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.RemoteAddr = "10.0.0.1:5678"

	key := KeyByUserID(c)
	if key != "ip:10.0.0.1" {
		t.Errorf("KeyByUserID = %q, want %q", key, "ip:10.0.0.1")
	}
}

func TestKeyByUserID_EmptyString_FallsBackToIP(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.RemoteAddr = "10.0.0.1:5678"
	c.Set(utils.ContextKeyUserID, "")

	key := KeyByUserID(c)
	if key != "ip:10.0.0.1" {
		t.Errorf("KeyByUserID = %q, want %q", key, "ip:10.0.0.1")
	}
}

// --- Config defaults tests ---

func TestApplyDefaults_ZeroValues(t *testing.T) {
	cfg := RateLimiterConfig{}
	cfg.applyDefaults()

	if cfg.CleanupInterval != 2*time.Minute {
		t.Errorf("CleanupInterval = %v, want %v", cfg.CleanupInterval, 2*time.Minute)
	}
	if cfg.IdleTTL != 10*time.Minute {
		t.Errorf("IdleTTL = %v, want %v", cfg.IdleTTL, 10*time.Minute)
	}
	if cfg.KeyFunc == nil {
		t.Error("KeyFunc should default to non-nil")
	}
}

func TestApplyDefaults_IdleTTLLessThanCleanupInterval(t *testing.T) {
	cfg := RateLimiterConfig{
		CleanupInterval: 10 * time.Minute,
		IdleTTL:         5 * time.Second,
	}
	cfg.applyDefaults()

	if cfg.IdleTTL != cfg.CleanupInterval {
		t.Errorf("IdleTTL = %v, want %v (should match CleanupInterval)", cfg.IdleTTL, cfg.CleanupInterval)
	}
}

// --- Token bucket tests ---

func TestAllow_WithinBurst(t *testing.T) {
	rl := &rateLimiter{
		entries: make(map[string]*limiterEntry),
		config:  RateLimiterConfig{Rate: 1, Burst: 5},
	}

	for i := 0; i < 5; i++ {
		if !rl.allow("key") {
			t.Errorf("request %d should be allowed within burst", i+1)
		}
	}
}

func TestAllow_ExceedsBurst(t *testing.T) {
	rl := &rateLimiter{
		entries: make(map[string]*limiterEntry),
		config:  RateLimiterConfig{Rate: 1, Burst: 3},
	}

	for i := 0; i < 3; i++ {
		rl.allow("key")
	}

	if rl.allow("key") {
		t.Error("request exceeding burst should be rejected")
	}
}

func TestAllow_BurstBoundary_ExactLimit(t *testing.T) {
	rl := &rateLimiter{
		entries: make(map[string]*limiterEntry),
		config:  RateLimiterConfig{Rate: 1, Burst: 3},
	}

	// Use exactly burst count
	for i := 0; i < 3; i++ {
		if !rl.allow("key") {
			t.Fatalf("request %d should be allowed (within burst of 3)", i+1)
		}
	}

	// Next request must be rejected
	if rl.allow("key") {
		t.Error("request at burst+1 should be rejected")
	}

	// Simulate token refill: 1 token at rate=1/sec after 1 second
	rl.mu.Lock()
	rl.entries["key"].lastCheck = rl.entries["key"].lastCheck.Add(-1 * time.Second)
	rl.mu.Unlock()

	// Should recover with 1 token
	if !rl.allow("key") {
		t.Error("request should be allowed after token refill")
	}

	// But only 1 token was refilled
	if rl.allow("key") {
		t.Error("second request after refill should be rejected (only 1 token refilled)")
	}
}

func TestAllow_TokenRefill(t *testing.T) {
	rl := &rateLimiter{
		entries: make(map[string]*limiterEntry),
		config:  RateLimiterConfig{Rate: 10, Burst: 5},
	}

	// Exhaust burst
	for i := 0; i < 5; i++ {
		rl.allow("key")
	}
	if rl.allow("key") {
		t.Fatal("should be exhausted")
	}

	// Simulate 0.5s passing at rate=10/s => 5 tokens refilled (capped at burst)
	rl.mu.Lock()
	rl.entries["key"].lastCheck = rl.entries["key"].lastCheck.Add(-500 * time.Millisecond)
	rl.mu.Unlock()

	for i := 0; i < 5; i++ {
		if !rl.allow("key") {
			t.Errorf("request %d should be allowed after refill", i+1)
		}
	}
}

func TestAllow_SeparateKeys(t *testing.T) {
	rl := &rateLimiter{
		entries: make(map[string]*limiterEntry),
		config:  RateLimiterConfig{Rate: 1, Burst: 1},
	}

	if !rl.allow("user:1") {
		t.Error("first key should be allowed")
	}
	if !rl.allow("user:2") {
		t.Error("second key should be allowed (separate bucket)")
	}
	if rl.allow("user:1") {
		t.Error("first key should be rejected (exhausted)")
	}
}

// --- Eviction tests ---

func TestEvict_RemovesIdleEntries(t *testing.T) {
	cfg := RateLimiterConfig{
		Rate:            1,
		Burst:           5,
		CleanupInterval: time.Minute,
		IdleTTL:         5 * time.Minute,
	}
	cfg.applyDefaults()

	rl := &rateLimiter{
		entries: make(map[string]*limiterEntry),
		config:  cfg,
	}

	// Add an entry that's been idle longer than IdleTTL
	rl.entries["old"] = &limiterEntry{
		tokens:    5,
		lastCheck: time.Now().Add(-10 * time.Minute),
	}
	// Add a recent entry
	rl.entries["recent"] = &limiterEntry{
		tokens:    5,
		lastCheck: time.Now(),
	}

	rl.evict()

	if _, exists := rl.entries["old"]; exists {
		t.Error("old entry should have been evicted")
	}
	if _, exists := rl.entries["recent"]; !exists {
		t.Error("recent entry should still exist")
	}
}

// --- Middleware integration tests ---

func TestRateLimit_Middleware_RejectsOverBurst(t *testing.T) {
	handler := RateLimit(RateLimiterConfig{
		Rate:     1,
		Burst:    2,
		TierName: "test",
		KeyFunc:  KeyByIP,
	})

	statuses := make([]int, 4)
	for i := 0; i < 4; i++ {
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		r.Use(handler)
		r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })
		c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
		c.Request.RemoteAddr = "1.2.3.4:1234"
		r.ServeHTTP(w, c.Request)
		statuses[i] = w.Code
	}

	// First 2 should be 200, rest should be 429
	for i := 0; i < 2; i++ {
		if statuses[i] != http.StatusOK {
			t.Errorf("request %d: got %d, want %d", i+1, statuses[i], http.StatusOK)
		}
	}
	for i := 2; i < 4; i++ {
		if statuses[i] != http.StatusTooManyRequests {
			t.Errorf("request %d: got %d, want %d", i+1, statuses[i], http.StatusTooManyRequests)
		}
	}
}

func TestRateLimit_Middleware_MetricsTierLabel(t *testing.T) {
	// Reset the metric to get a clean baseline
	metrics.RateLimitRejectsTotal.Reset()

	handler := RateLimit(RateLimiterConfig{
		Rate:     1,
		Burst:    1,
		TierName: "strict",
		KeyFunc:  KeyByIP,
	})

	// First request succeeds, second triggers rejection with metric
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)
		r.Use(handler)
		r.GET("/api/v1/auth/login", func(c *gin.Context) { c.Status(http.StatusOK) })
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/login", nil)
		req.RemoteAddr = "5.6.7.8:1234"
		r.ServeHTTP(w, req)
	}

	// Verify the metric was recorded with the correct tier label
	m, err := metrics.RateLimitRejectsTotal.GetMetricWithLabelValues("/api/v1/auth/login", "strict")
	if err != nil {
		t.Fatalf("failed to get metric: %v", err)
	}

	// Read the counter value via the Write method
	var pm dto.Metric
	if err := m.Write(&pm); err != nil {
		t.Fatalf("failed to write metric: %v", err)
	}
	if pm.GetCounter().GetValue() != 1 {
		t.Errorf("expected 1 rejection, got %v", pm.GetCounter().GetValue())
	}
}
