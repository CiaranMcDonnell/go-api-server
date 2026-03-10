package middleware

import (
	"net/http"

	"github.com/ciaranmcdonnell/go-api-server/pkg/apperrors"
	"github.com/gin-gonic/gin"
)

// BodyLimit returns a Gin middleware that rejects requests with a body
// larger than maxBytes. This prevents OOM from oversized payloads.
func BodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}

// HandleBodyTooLarge is a helper to detect MaxBytesReader errors in handlers.
// Call this when BindJSON returns an error to provide a clear message.
func IsBodyTooLarge(err error) bool {
	return err != nil && err.Error() == "http: request body too large"
}

// BodyTooLargeResponse sends a standardized 413 response.
func BodyTooLargeResponse(c *gin.Context) {
	apperrors.Error(c, http.StatusRequestEntityTooLarge, "body_too_large", "Request body too large")
}
