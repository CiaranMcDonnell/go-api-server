package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/ciaranmcdonnell/go-api-server/pkg/apperrors"
	"github.com/gin-gonic/gin"
)

// Timeout returns a Gin middleware that enforces a per-request context deadline.
// If the handler does not complete within the timeout, the context is cancelled
// so database queries and downstream calls abort early.
func Timeout(duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), duration)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		if ctx.Err() == context.DeadlineExceeded {
			apperrors.Error(c, http.StatusGatewayTimeout, "timeout", "Request timed out")
			c.Abort()
		}
	}
}
