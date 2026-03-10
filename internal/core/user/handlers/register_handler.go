package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	userservice "github.com/ciaranmcdonnell/go-api-server/internal/core/user/service"
	"github.com/ciaranmcdonnell/go-api-server/models"
	"github.com/ciaranmcdonnell/go-api-server/pkg/apperrors"
	"github.com/ciaranmcdonnell/go-api-server/pkg/utils"
)

func RegisterHandler(userService userservice.UserServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateUserRequest
		if err := utils.BindJSON(c.Request.Body, &req); err != nil {
			apperrors.Error(c, http.StatusBadRequest, "validation_error", err.Error())
			return
		}

		id, err := userService.RegisterUser(c.Request.Context(), &req)
		if err != nil {
			if errors.Is(err, apperrors.ErrConflict) {
				apperrors.Error(c, http.StatusConflict, "email_taken", "Email already registered")
				return
			}
			apperrors.Error(c, http.StatusInternalServerError, "internal_error", "Failed to create user")
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "user_id": id})
	}
}
