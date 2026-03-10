package routes

import (
	"github.com/gin-gonic/gin"
	groups "github.com/ciaranmcdonnell/go-api-server/api/v1/routes/groups"
	auditMiddleware "github.com/ciaranmcdonnell/go-api-server/internal/core/audit/interfaces/http/service"
	commonHandlers "github.com/ciaranmcdonnell/go-api-server/internal/core/common/handlers"
	commonService "github.com/ciaranmcdonnell/go-api-server/internal/core/common/service"
)

func RegisterRoutes(router *gin.Engine, services commonService.ServicesInterface, auditHandler auditMiddleware.AuditMiddleware) {

	router.GET("/health", commonHandlers.HealthCheckHandler(services))

	authService := services.GetAuthService()
	userService := services.GetUserService()
	config := services.GetConfig()

	itemService := services.GetItemService()

	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			groups.AuthRoutes(v1, authService, userService, auditHandler, config)
			groups.ItemRoutes(v1, authService, itemService, auditHandler)
		}
	}
}
