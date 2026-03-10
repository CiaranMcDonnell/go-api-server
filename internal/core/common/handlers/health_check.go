package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ciaranmcdonnell/go-api-server/internal/core/common/service"
	"github.com/ciaranmcdonnell/go-api-server/internal/database"
)

var startTime = time.Now()

func HealthCheckHandler(servicesManager service.ServicesInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		healthStatus := gin.H{
			"status":    "ok",
			"version":   "1.0.0",
			"timestamp": time.Now().Format(time.RFC3339),
			"uptime":    time.Since(startTime).String(),
			"services":  make(map[string]string),
			"runtime": gin.H{
				"go_version": runtime.Version(),
				"goroutines": runtime.NumGoroutine(),
				"arch":       runtime.GOARCH,
				"os":         runtime.GOOS,
			},
		}

		dbStatus := "ok"
		err := database.HealthCheck(c.Request.Context())
		if err != nil {
			dbStatus = "error: " + err.Error()
			healthStatus["status"] = "degraded"
		}
		healthStatus["services"].(map[string]string)["database"] = dbStatus

		statusCode := http.StatusOK
		if healthStatus["status"] != "ok" {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, healthStatus)
	}
}
