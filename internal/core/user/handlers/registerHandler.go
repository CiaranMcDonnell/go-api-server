package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ciaranmcdonnell/go-api-server/models"
	userservice "github.com/ciaranmcdonnell/go-api-server/internal/core/user/service"
	"github.com/ciaranmcdonnell/go-api-server/pkg/utils"
)

func RegisterHandler(userService userservice.UserServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User

		if ok, errMsg := utils.ParseJSONBody(c.Request.Body, &user); !ok {
			slog.Warn("Failed to parse registration body", "error", errMsg)
			c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
			return
		}

		if user.Name == "" || user.Password == "" || user.Email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email, name and password are required"})
			return
		}

		if !utils.NameRegex.MatchString(user.Name) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid name format"})
			return
		}

		if !utils.EmailRegex.MatchString(user.Email) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
			return
		}

		if !utils.PasswordRegex.MatchString(user.Password) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be between 8 and 128 characters"})
			return
		}

		id, err := userService.RegisterUser(c.Request.Context(), &user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "user_id": id})
	}
}
