package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	authservice "github.com/ciaranmcdonnell/go-api-server/internal/core/auth/service"
	"github.com/ciaranmcdonnell/go-api-server/pkg/utils"
)

func LoginHandler(authService authservice.AuthServiceInterface, config *utils.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginData struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if ok, errMsg := utils.ParseJSONBody(c.Request.Body, &loginData); !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
			return
		}

		if loginData.Password == "" || loginData.Email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email and password are required"})
			return
		}

		user, err := authService.AuthenticateUser(c.Request.Context(), loginData.Email, loginData.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		token, err := authService.GenerateAuthToken(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		maxAge := config.CookieMaxAgeSecs
		if maxAge <= 0 {
			maxAge = config.JWTExpirationHours * 3600
		}
		secure := config.Environment != "development"
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie(utils.CookieName, token, maxAge, "/", "", secure, true)
		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

func LogoutHandler(authService authservice.AuthServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie(utils.CookieName, "", -1, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
	}
}

func MeHandler(authService authservice.AuthServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDValue, exists := c.Get(utils.ContextKeyUserID)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
			return
		}

		userIDStr, ok := userIDValue.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
			return
		}

		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user ID"})
			return
		}

		user, err := authService.GetCurrentUser(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"user": user})
	}
}
