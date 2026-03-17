package groups

import (
	"github.com/gin-gonic/gin"
	auditMiddleware "github.com/ciaranmcdonnell/go-api-server/internal/core/audit/interfaces/http/service"
	authHandlers "github.com/ciaranmcdonnell/go-api-server/internal/core/auth/handlers"
	authmid "github.com/ciaranmcdonnell/go-api-server/internal/core/auth/middleware"
	authsvc "github.com/ciaranmcdonnell/go-api-server/internal/core/auth/service"
	commonMiddleware "github.com/ciaranmcdonnell/go-api-server/internal/core/common/middleware"
	userHandlers "github.com/ciaranmcdonnell/go-api-server/internal/core/user/handlers"
	usersvc "github.com/ciaranmcdonnell/go-api-server/internal/core/user/service"
	"github.com/ciaranmcdonnell/go-api-server/pkg/utils"
)

func AuthRoutes(router *gin.RouterGroup, authService authsvc.AuthServiceInterface, userService usersvc.UserServiceInterface, auditService auditMiddleware.AuditMiddleware, config *utils.Config) {
	strictLimiter := commonMiddleware.RateLimit(commonMiddleware.RateLimiterConfig{
		Rate:     config.RateLimitStrictRate,
		Burst:    config.RateLimitStrictBurst,
		KeyFunc:  commonMiddleware.KeyByIP,
		TierName: "strict",
	})
	standardLimiter := commonMiddleware.RateLimit(commonMiddleware.RateLimiterConfig{
		Rate:     config.RateLimitStandardRate,
		Burst:    config.RateLimitStandardBurst,
		KeyFunc:  commonMiddleware.KeyByUserID,
		TierName: "standard",
	})
	relaxedLimiter := commonMiddleware.RateLimit(commonMiddleware.RateLimiterConfig{
		Rate:     config.RateLimitRelaxedRate,
		Burst:    config.RateLimitRelaxedBurst,
		KeyFunc:  commonMiddleware.KeyByUserID,
		TierName: "relaxed",
	})

	audit := func(c *gin.Context) { auditService(c) }

	// Unauthenticated — Strict tier, keyed by IP
	publicAuth := router.Group("/auth")
	publicAuth.Use(strictLimiter)
	{
		publicAuth.POST("/login", audit, authHandlers.LoginHandler(authService, config))
		publicAuth.POST("/register", audit, userHandlers.RegisterHandler(userService))
	}

	// Authenticated — RequireAuth first, then per-endpoint tiers
	protectedAuth := router.Group("/auth")
	protectedAuth.Use(authmid.RequireAuth(authService))
	{
		protectedAuth.GET("/me", relaxedLimiter, authHandlers.MeHandler(authService))
		protectedAuth.POST("/logout", standardLimiter, audit, authHandlers.LogoutHandler(authService))
	}
}
