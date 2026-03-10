package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	authservice "github.com/ciaranmcdonnell/go-api-server/internal/core/auth/service"
	"github.com/ciaranmcdonnell/go-api-server/internal/metrics"
	"github.com/ciaranmcdonnell/go-api-server/models"
	"github.com/ciaranmcdonnell/go-api-server/pkg/cache"
	"github.com/ciaranmcdonnell/go-api-server/pkg/utils"
)

var tokenCache = cache.New[string, *models.Claims](30 * time.Second)

func RequireAuth(authService authservice.AuthServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := authservice.ExtractTokenFromRequest(
			c.GetHeader("Authorization"),
			getCookie(c, utils.CookieName),
		)

		// Check cache first
		if claims, ok := tokenCache.Get(tokenString); ok {
			metrics.CacheHitsTotal.WithLabelValues("token").Inc()
			c.Set(utils.ContextKeyUserID, claims.UserID)
			c.Next()
			return
		}
		metrics.CacheMissesTotal.WithLabelValues("token").Inc()

		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Cache with TTL matching remaining JWT lifetime
		ttl := time.Until(claims.ExpiresAt.Time)
		if ttl > 0 {
			tokenCache.Set(tokenString, claims, ttl)
		}

		c.Set(utils.ContextKeyUserID, claims.UserID)
		c.Next()
	}
}

func getCookie(c *gin.Context, name string) string {
	val, err := c.Cookie(name)
	if err != nil {
		return ""
	}
	return val
}
