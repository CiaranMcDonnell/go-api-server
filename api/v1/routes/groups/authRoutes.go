package groups

import (
	"github.com/gin-gonic/gin"
	auditMiddleware "github.com/ciaranmcdonnell/go-api-server/internal/core/audit/interfaces/http/service"
	authHandlers "github.com/ciaranmcdonnell/go-api-server/internal/core/auth/handlers"
	authmid "github.com/ciaranmcdonnell/go-api-server/internal/core/auth/middleware"
	authsvc "github.com/ciaranmcdonnell/go-api-server/internal/core/auth/service"
	userHandlers "github.com/ciaranmcdonnell/go-api-server/internal/core/user/handlers"
	usersvc "github.com/ciaranmcdonnell/go-api-server/internal/core/user/service"
	"github.com/ciaranmcdonnell/go-api-server/pkg/utils"
)

func AuthRoutes(router *gin.RouterGroup, authService authsvc.AuthServiceInterface, userService usersvc.UserServiceInterface, auditService auditMiddleware.AuditMiddleware, config *utils.Config) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", func(c *gin.Context) { auditService(c) }, authHandlers.LoginHandler(authService, config))
		registerGroup := authGroup.Group("/register")
		registerGroup.Use(func(c *gin.Context) { auditService(c) })
		{
			registerGroup.POST("", userHandlers.RegisterHandler(userService))
		}

		protected := authGroup.Group("")
		protected.Use(authmid.RequireAuth(authService))
		{
			protected.GET("/me", authHandlers.MeHandler(authService))
			protected.POST("/logout", func(c *gin.Context) { auditService(c) }, authHandlers.LogoutHandler(authService))
		}
	}
}
