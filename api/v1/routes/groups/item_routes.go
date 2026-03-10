package groups

import (
	"github.com/gin-gonic/gin"
	auditMiddleware "github.com/ciaranmcdonnell/go-api-server/internal/core/audit/interfaces/http/service"
	authmid "github.com/ciaranmcdonnell/go-api-server/internal/core/auth/middleware"
	authsvc "github.com/ciaranmcdonnell/go-api-server/internal/core/auth/service"
	itemHandlers "github.com/ciaranmcdonnell/go-api-server/internal/core/items/handlers"
	itemInterfaces "github.com/ciaranmcdonnell/go-api-server/internal/core/items/domain/interfaces"
)

func ItemRoutes(router *gin.RouterGroup, authService authsvc.AuthServiceInterface, itemService itemInterfaces.ItemService, auditHandler auditMiddleware.AuditMiddleware) {
	items := router.Group("/items")
	items.Use(authmid.RequireAuth(authService))
	items.Use(func(c *gin.Context) { auditHandler(c) })
	{
		items.POST("", itemHandlers.CreateHandler(itemService))
		items.GET("", itemHandlers.ListHandler(itemService))
		items.GET("/:id", itemHandlers.GetHandler(itemService))
		items.PUT("/:id", itemHandlers.UpdateHandler(itemService))
		items.DELETE("/:id", itemHandlers.DeleteHandler(itemService))
	}
}
