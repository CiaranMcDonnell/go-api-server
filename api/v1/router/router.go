package router

import (
	"log/slog"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/ciaranmcdonnell/go-api-server/api/v1/routes"
	auditService "github.com/ciaranmcdonnell/go-api-server/internal/core/audit/application/service"
	auditMiddleware "github.com/ciaranmcdonnell/go-api-server/internal/core/audit/interfaces/http/service"
	"github.com/ciaranmcdonnell/go-api-server/internal/core/audit/worker"
	commonMiddleware "github.com/ciaranmcdonnell/go-api-server/internal/core/common/middleware"
	repository "github.com/ciaranmcdonnell/go-api-server/internal/core/common/repository"
	commonService "github.com/ciaranmcdonnell/go-api-server/internal/core/common/service"
	"github.com/ciaranmcdonnell/go-api-server/internal/database"
	"github.com/ciaranmcdonnell/go-api-server/internal/metrics"
	"github.com/ciaranmcdonnell/go-api-server/pkg/utils"
)

func Setup(config *utils.Config, servicesManager commonService.ServicesInterface, queriesManager repository.QueriesInterface) (*gin.Engine, *worker.Pool) {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(metrics.PrometheusMiddleware())
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	if database.DBPool != nil {
		prometheus.MustRegister(metrics.NewDBPoolCollector(database.DBPool))
	}

	r.Use(commonMiddleware.RequestID())

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

	auditRepo := queriesManager.GetAuditQueries()
	auditSvc := auditService.NewAuditService(auditRepo)
	auditPool := worker.NewPool(10, 1000, auditSvc)

	auditMiddlewareConfig := auditMiddleware.AuditMiddlewareConfig{
		SkipPaths:  []string{"/health", "/metrics"},
		WorkerPool: auditPool,
	}
	auditHandler := auditMiddleware.NewAuditMiddleware(auditMiddlewareConfig)
	routes.RegisterRoutes(r, servicesManager, auditHandler)

	slog.Info("Routes registered successfully")

	return r, auditPool
}
