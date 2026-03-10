package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	authservice "github.com/ciaranmcdonnell/go-api-server/internal/core/auth/service"
	"github.com/ciaranmcdonnell/go-api-server/models"
	"github.com/ciaranmcdonnell/go-api-server/pkg/apperrors"
	"github.com/ciaranmcdonnell/go-api-server/pkg/utils"
)

func LoginHandler(authService authservice.AuthServiceInterface, config *utils.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.LoginRequest
		if err := utils.BindJSON(c.Request.Body, &req); err != nil {
			apperrors.Error(c, http.StatusBadRequest, "validation_error", err.Error())
			return
		}

		user, err := authService.AuthenticateUser(c.Request.Context(), req.Email, req.Password)
		if err != nil {
			apperrors.Error(c, http.StatusUnauthorized, "invalid_credentials", "Invalid credentials")
			return
		}

		token, err := authService.GenerateAuthToken(user)
		if err != nil {
			apperrors.Error(c, http.StatusInternalServerError, "internal_error", "Internal server error")
			return
		}

		maxAge := config.CookieMaxAgeSecs
		if maxAge <= 0 {
			maxAge = config.JWTExpirationHours * 3600
		}
		secure := config.Environment != "development"
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie(utils.CookieName, token, maxAge, "/", "", secure, true)
		c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
	}
}

func LogoutHandler(authService authservice.AuthServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie(utils.CookieName, "", -1, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
	}
}

func MeHandler(authService authservice.AuthServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDValue, exists := c.Get(utils.ContextKeyUserID)
		if !exists {
			apperrors.Error(c, http.StatusUnauthorized, "unauthorized", "Authentication required")
			return
		}

		userIDStr, ok := userIDValue.(string)
		if !ok {
			apperrors.Error(c, http.StatusInternalServerError, "internal_error", "Invalid user ID")
			return
		}

		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			apperrors.Error(c, http.StatusInternalServerError, "internal_error", "Invalid user ID")
			return
		}

		user, err := authService.GetCurrentUser(c.Request.Context(), userID)
		if err != nil {
			apperrors.Error(c, http.StatusInternalServerError, "internal_error", "Failed to fetch user")
			return
		}

		c.JSON(http.StatusOK, gin.H{"user": user.ToResponse()})
	}
}
