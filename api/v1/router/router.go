package router

import (
	"log/slog"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/ciaranmcdonnell/go-api-server/api/v1/routes"
	auditService "github.com/ciaranmcdonnell/go-api-server/internal/core/audit/application/service"
	auditMiddleware "github.com/ciaranmcdonnell/go-api-server/internal/core/audit/interfaces/http/service"
	commonMiddleware "github.com/ciaranmcdonnell/go-api-server/internal/core/common/middleware"
	repository "github.com/ciaranmcdonnell/go-api-server/internal/core/common/repository"
	commonService "github.com/ciaranmcdonnell/go-api-server/internal/core/common/service"
	"github.com/ciaranmcdonnell/go-api-server/pkg/utils"
)

func Setup(config *utils.Config, servicesManager commonService.ServicesInterface, queriesManager repository.QueriesInterface) *gin.Engine {
	r := gin.Default()

	// Request ID middleware
	r.Use(commonMiddleware.RequestID())

	// CORS setup
	corsConfig := cors.DefaultConfig()
	if config.CORSOrigins != "" {
		corsConfig.AllowOrigins = strings.Split(config.CORSOrigins, ",")
	} else {
		slog.Warn("CORS_ORIGINS not set, defaulting to localhost only")
		corsConfig.AllowOrigins = []string{"http://localhost:3000"}
	}
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	r.Use(cors.New(corsConfig))

	// Setup Audit Service & Middleware
	auditRepo := queriesManager.GetAuditQueries()
	auditSvc := auditService.NewAuditService(auditRepo)

	auditMiddlewareConfig := auditMiddleware.AuditMiddlewareConfig{
		SkipPaths: []string{"/health"},
		Service:   auditSvc,
	}
	auditHandler := auditMiddleware.NewAuditMiddleware(auditMiddlewareConfig)
	routes.RegisterRoutes(r, servicesManager, auditHandler)

	slog.Info("Routes registered successfully")

	return r
}
