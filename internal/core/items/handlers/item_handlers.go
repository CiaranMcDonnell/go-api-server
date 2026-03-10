package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ciaranmcdonnell/go-api-server/internal/core/items/domain/interfaces"
	"github.com/ciaranmcdonnell/go-api-server/internal/core/items/domain/models"
	sharedmodels "github.com/ciaranmcdonnell/go-api-server/models"
	"github.com/ciaranmcdonnell/go-api-server/pkg/apperrors"
	"github.com/ciaranmcdonnell/go-api-server/pkg/utils"
)

func getUserID(c *gin.Context) (int64, bool) {
	val, exists := c.Get(utils.ContextKeyUserID)
	if !exists {
		apperrors.Error(c, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return 0, false
	}
	str, ok := val.(string)
	if !ok {
		apperrors.Error(c, http.StatusInternalServerError, "internal_error", "Invalid user ID")
		return 0, false
	}
	id, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		apperrors.Error(c, http.StatusInternalServerError, "internal_error", "Invalid user ID")
		return 0, false
	}
	return id, true
}

func CreateHandler(svc interfaces.ItemService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getUserID(c)
		if !ok {
			return
		}

		var dto models.CreateItemDTO
		if err := utils.BindJSON(c.Request.Body, &dto); err != nil {
			apperrors.Error(c, http.StatusBadRequest, "validation_error", err.Error())
			return
		}

		item, err := svc.CreateItem(c.Request.Context(), userID, dto)
		if err != nil {
			apperrors.MapError(c, err)
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
			apperrors.Error(c, http.StatusBadRequest, "invalid_id", "Invalid item ID")
			return
		}

		item, err := svc.GetItem(c.Request.Context(), userID, itemID)
		if err != nil {
			apperrors.MapError(c, err)
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

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
		params := sharedmodels.NewPaginationParams(page, perPage)

		items, pagination, err := svc.ListItems(c.Request.Context(), userID, params)
		if err != nil {
			apperrors.MapError(c, err)
			return
		}

		if items == nil {
			items = []*models.Item{}
		}

		c.JSON(http.StatusOK, gin.H{
			"items":      models.ItemsToResponses(items),
			"pagination": pagination,
		})
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
			apperrors.Error(c, http.StatusBadRequest, "invalid_id", "Invalid item ID")
			return
		}

		var dto models.UpdateItemDTO
		if err := utils.BindJSON(c.Request.Body, &dto); err != nil {
			apperrors.Error(c, http.StatusBadRequest, "validation_error", err.Error())
			return
		}

		item, err := svc.UpdateItem(c.Request.Context(), userID, itemID, dto)
		if err != nil {
			apperrors.MapError(c, err)
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
			apperrors.Error(c, http.StatusBadRequest, "invalid_id", "Invalid item ID")
			return
		}

		if err := svc.DeleteItem(c.Request.Context(), userID, itemID); err != nil {
			apperrors.MapError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Item deleted successfully"})
	}
}
