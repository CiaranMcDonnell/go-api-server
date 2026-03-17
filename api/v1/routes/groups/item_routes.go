package groups

import (
	"github.com/gin-gonic/gin"
	auditMiddleware "github.com/ciaranmcdonnell/go-api-server/internal/core/audit/interfaces/http/service"
	authmid "github.com/ciaranmcdonnell/go-api-server/internal/core/auth/middleware"
	authsvc "github.com/ciaranmcdonnell/go-api-server/internal/core/auth/service"
	commonMiddleware "github.com/ciaranmcdonnell/go-api-server/internal/core/common/middleware"
	itemHandlers "github.com/ciaranmcdonnell/go-api-server/internal/core/items/handlers"
	itemInterfaces "github.com/ciaranmcdonnell/go-api-server/internal/core/items/domain/interfaces"
	"github.com/ciaranmcdonnell/go-api-server/pkg/utils"
)

func ItemRoutes(router *gin.RouterGroup, authService authsvc.AuthServiceInterface, itemService itemInterfaces.ItemService, auditHandler auditMiddleware.AuditMiddleware, config *utils.Config) {
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

	items := router.Group("/items")
	items.Use(authmid.RequireAuth(authService))
	{
		audit := func(c *gin.Context) { auditHandler(c) }

		// Reads — Relaxed tier (list + hydrate causes bursts)
		items.GET("", relaxedLimiter, audit, itemHandlers.ListHandler(itemService))
		items.GET("/:id", relaxedLimiter, audit, itemHandlers.GetHandler(itemService))

		// Mutations — Standard tier
		items.POST("", standardLimiter, audit, itemHandlers.CreateHandler(itemService))
		items.PUT("/:id", standardLimiter, audit, itemHandlers.UpdateHandler(itemService))
		items.DELETE("/:id", standardLimiter, audit, itemHandlers.DeleteHandler(itemService))
	}
}
