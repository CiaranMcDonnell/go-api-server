package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ciaranmcdonnell/go-api-server/models"
	userservice "github.com/ciaranmcdonnell/go-api-server/internal/core/user/service"
	"github.com/ciaranmcdonnell/go-api-server/pkg/utils"
)

func RegisterHandler(userService userservice.UserServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateUserRequest
		if err := utils.BindJSON(c.Request.Body, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		id, err := userService.RegisterUser(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "user_id": id})
	}
}
