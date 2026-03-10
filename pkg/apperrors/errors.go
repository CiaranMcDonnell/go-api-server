package apperrors

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	ErrNotFound           = errors.New("not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrConflict           = errors.New("conflict")
	ErrBadRequest         = errors.New("bad request")
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func Error(c *gin.Context, status int, code string, message string) {
	c.JSON(status, gin.H{"error": ErrorResponse{Code: code, Message: message}})
}

func MapError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		Error(c, http.StatusNotFound, "not_found", "Resource not found")
	case errors.Is(err, ErrUnauthorized):
		Error(c, http.StatusUnauthorized, "unauthorized", "Authentication required")
	case errors.Is(err, ErrInvalidCredentials):
		Error(c, http.StatusUnauthorized, "invalid_credentials", "Invalid credentials")
	case errors.Is(err, ErrConflict):
		Error(c, http.StatusConflict, "conflict", "Resource already exists")
	case errors.Is(err, ErrBadRequest):
		Error(c, http.StatusBadRequest, "bad_request", err.Error())
	default:
		Error(c, http.StatusInternalServerError, "internal_error", "Internal server error")
	}
}
