package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	auditDomainInterfaces "github.com/ciaranmcdonnell/go-api-server/internal/core/audit/domain/interfaces"
	"github.com/ciaranmcdonnell/go-api-server/internal/core/audit/domain/models"
	"github.com/ciaranmcdonnell/go-api-server/pkg/utils"
)

type AuditMiddleware gin.HandlerFunc

type AuditMiddlewareConfig struct {
	SkipPaths []string
	Service   auditDomainInterfaces.AuditService
}

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func NewAuditMiddleware(config AuditMiddlewareConfig) AuditMiddleware {
	return func(c *gin.Context) {
		startTime := time.Now()

		for _, path := range config.SkipPaths {
			if strings.HasPrefix(c.Request.URL.Path, path) {
				c.Next()
				return
			}
		}

		var requestBodyBytes []byte
		if c.Request.Body != nil {
			var err error
			requestBodyBytes, err = io.ReadAll(c.Request.Body)
			if err != nil {
				slog.Warn("Failed to read request body for audit", "error", err)
			} else {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBodyBytes))
			}
		}

		responseWriter := &responseBodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = responseWriter

		c.Next()

		statusCode := responseWriter.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		resource := extractResource(path)

		var userID string = "anonymous"
		var attemptedIdentifier string
		var requestBodyToLog string = string(requestBodyBytes)

		if strings.HasSuffix(path, "/login") || strings.HasSuffix(path, "/register") {
			requestBodyToLog = "[REDACTED]"

			var authPayload struct {
				Email    string `json:"email"`
				Username string `json:"username"`
			}
			if err := json.Unmarshal(requestBodyBytes, &authPayload); err == nil {
				if authPayload.Email != "" {
					attemptedIdentifier = authPayload.Email
				} else if authPayload.Username != "" && strings.HasSuffix(path, "/login") {
					attemptedIdentifier = authPayload.Username
				}
			}
		} else {
			if userIDVal, exists := c.Get(utils.ContextKeyUserID); exists {
				if uid, ok := userIDVal.(string); ok && uid != "" {
					userID = uid
				}
			}
		}

		dto := models.CreateAuditLogDTO{
			UserID:              userID,
			AttemptedIdentifier: attemptedIdentifier,
			Action:              determineAction(method),
			Resource:            resource,
			RequestPath:         path,
			Method:              method,
			StatusCode:          statusCode,
			IPAddress:           c.ClientIP(),
			UserAgent:           c.Request.UserAgent(),
			RequestBody:         requestBodyToLog,
		}

		go func() {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("Panic in audit logging", "recover", r)
				}
			}()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := config.Service.LogAuditEvent(ctx, dto); err != nil {
				slog.Error("Failed to log audit event", "error", err)
			}
		}()

		latency := time.Since(startTime)
		slog.Info("Request processed",
			"user", userID,
			"path", path,
			"method", method,
			"status", statusCode,
			"latency", latency.String(),
		)
	}
}

func determineAction(method string) string {
	switch method {
	case http.MethodPost:
		return "CREATE"
	case http.MethodPut:
		return "UPDATE"
	case http.MethodPatch:
		return "UPDATE"
	case http.MethodDelete:
		return "DELETE"
	case http.MethodGet:
		return "READ"
	default:
		return strings.ToUpper(method)
	}
}

func extractResource(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 {
		return parts[2]
	}
	if len(parts) >= 1 {
		return parts[0]
	}
	return "unknown"
}
