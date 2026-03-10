package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ciaranmcdonnell/go-api-server/internal/core/items/domain/interfaces"
	"github.com/ciaranmcdonnell/go-api-server/internal/core/items/domain/models"
	"github.com/ciaranmcdonnell/go-api-server/pkg/apperrors"
	"github.com/ciaranmcdonnell/go-api-server/pkg/utils"
)

func getUserID(c *gin.Context) (int64, bool) {
	val, exists := c.Get(utils.ContextKeyUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return 0, false
	}
	str, ok := val.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return 0, false
	}
	id, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return 0, false
	}
	return id, true
}

func errorResponse(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
	case errors.Is(err, apperrors.ErrConflict):
		c.JSON(http.StatusConflict, gin.H{"error": "Item already exists"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}

func CreateHandler(svc interfaces.ItemService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getUserID(c)
		if !ok {
			return
		}

		var dto models.CreateItemDTO
		if err := utils.BindJSON(c.Request.Body, &dto); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		item, err := svc.CreateItem(c.Request.Context(), userID, dto)
		if err != nil {
			errorResponse(c, err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{"item": item.ToResponse()})
	}
}

func GetHandler(svc interfaces.ItemService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getUserID(c)
		if !ok {
			return
		}

		itemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
			return
		}

		item, err := svc.GetItem(c.Request.Context(), userID, itemID)
		if err != nil {
			errorResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"item": item.ToResponse()})
	}
}

func ListHandler(svc interfaces.ItemService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getUserID(c)
		if !ok {
			return
		}

		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		items, err := svc.ListItems(c.Request.Context(), userID, limit, offset)
		if err != nil {
			errorResponse(c, err)
			return
		}

		if items == nil {
			items = []*models.Item{}
		}

		c.JSON(http.StatusOK, gin.H{"items": models.ItemsToResponses(items)})
	}
}

func UpdateHandler(svc interfaces.ItemService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getUserID(c)
		if !ok {
			return
		}

		itemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
			return
		}

		var dto models.UpdateItemDTO
		if err := utils.BindJSON(c.Request.Body, &dto); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		item, err := svc.UpdateItem(c.Request.Context(), userID, itemID, dto)
		if err != nil {
			errorResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"item": item.ToResponse()})
	}
}

func DeleteHandler(svc interfaces.ItemService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getUserID(c)
		if !ok {
			return
		}

		itemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
			return
		}

		if err := svc.DeleteItem(c.Request.Context(), userID, itemID); err != nil {
			errorResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Item deleted"})
	}
}
