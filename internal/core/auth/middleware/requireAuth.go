package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	authservice "github.com/ciaranmcdonnell/go-api-server/internal/core/auth/service"
	"github.com/ciaranmcdonnell/go-api-server/pkg/utils"
)

func RequireAuth(authService authservice.AuthServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := authservice.ExtractTokenFromRequest(
			c.GetHeader("Authorization"),
			getCookie(c, utils.CookieName),
		)

		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		c.Set(utils.ContextKeyUserID, claims.UserID)
		c.Next()
	}
}

func getCookie(c *gin.Context, name string) string {
	val, err := c.Cookie(name)
	if err != nil {
		return ""
	}
	return val
}
