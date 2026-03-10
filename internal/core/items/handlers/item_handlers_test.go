package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ciaranmcdonnell/go-api-server/internal/core/items/domain/models"
	"github.com/ciaranmcdonnell/go-api-server/internal/core/items/mocks"
	"github.com/ciaranmcdonnell/go-api-server/pkg/apperrors"
	"github.com/ciaranmcdonnell/go-api-server/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupRouter(svc *mocks.MockItemService) *gin.Engine {
	r := gin.New()
	items := r.Group("/items")
	items.Use(func(c *gin.Context) {
		c.Set(utils.ContextKeyUserID, "42")
		c.Next()
	})
	{
		items.POST("", CreateHandler(svc))
		items.GET("", ListHandler(svc))
		items.GET("/:id", GetHandler(svc))
		items.PUT("/:id", UpdateHandler(svc))
		items.DELETE("/:id", DeleteHandler(svc))
	}
	return r
}

func newTestItem(id, userID int64) *models.Item {
	return &models.Item{
		ID:          id,
		UserID:      userID,
		Name:        "Test Item",
		Description: "A test item",
		CreatedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func TestCreateHandler_Success(t *testing.T) {
	svc := new(mocks.MockItemService)
	router := setupRouter(svc)

	created := newTestItem(1, 42)
	svc.On("CreateItem", mock.Anything, int64(42), models.CreateItemDTO{Name: "New Item", Description: "desc"}).
		Return(created, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/items", strings.NewReader(`{"name":"New Item","description":"desc"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), `"name":"Test Item"`)
	svc.AssertExpectations(t)
}

func TestCreateHandler_ValidationError(t *testing.T) {
	svc := new(mocks.MockItemService)
	router := setupRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/items", strings.NewReader(`{"description":"no name"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "name is required")
	svc.AssertNotCalled(t, "CreateItem", mock.Anything, mock.Anything, mock.Anything)
}

func TestCreateHandler_InvalidJSON(t *testing.T) {
	svc := new(mocks.MockItemService)
	router := setupRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/items", strings.NewReader(`{broken`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid JSON")
}

func TestCreateHandler_ServiceError(t *testing.T) {
	svc := new(mocks.MockItemService)
	router := setupRouter(svc)

	svc.On("CreateItem", mock.Anything, int64(42), mock.Anything).
		Return(nil, fmt.Errorf("db error"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/items", strings.NewReader(`{"name":"Test"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetHandler_Success(t *testing.T) {
	svc := new(mocks.MockItemService)
	router := setupRouter(svc)

	item := newTestItem(1, 42)
	svc.On("GetItem", mock.Anything, int64(42), int64(1)).Return(item, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/items/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"name":"Test Item"`)
}

func TestGetHandler_NotFound(t *testing.T) {
	svc := new(mocks.MockItemService)
	router := setupRouter(svc)

	svc.On("GetItem", mock.Anything, int64(42), int64(999)).Return(nil, apperrors.ErrNotFound)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/items/999", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetHandler_InvalidID(t *testing.T) {
	svc := new(mocks.MockItemService)
	router := setupRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/items/abc", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid item ID")
}

func TestListHandler_Success(t *testing.T) {
	svc := new(mocks.MockItemService)
	router := setupRouter(svc)

	items := []*models.Item{newTestItem(1, 42), newTestItem(2, 42)}
	svc.On("ListItems", mock.Anything, int64(42), 20, 0).Return(items, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/items", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"items":[`)
}

func TestListHandler_Empty(t *testing.T) {
	svc := new(mocks.MockItemService)
	router := setupRouter(svc)

	svc.On("ListItems", mock.Anything, int64(42), 20, 0).Return(nil, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/items", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"items":[]`)
}

func TestListHandler_CustomPagination(t *testing.T) {
	svc := new(mocks.MockItemService)
	router := setupRouter(svc)

	svc.On("ListItems", mock.Anything, int64(42), 5, 10).Return([]*models.Item{}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/items?limit=5&offset=10", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestUpdateHandler_Success(t *testing.T) {
	svc := new(mocks.MockItemService)
	router := setupRouter(svc)

	updated := newTestItem(1, 42)
	updated.Name = "Updated"
	svc.On("UpdateItem", mock.Anything, int64(42), int64(1), mock.Anything).Return(updated, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/items/1", strings.NewReader(`{"name":"Updated"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"name":"Updated"`)
}

func TestUpdateHandler_NotFound(t *testing.T) {
	svc := new(mocks.MockItemService)
	router := setupRouter(svc)

	svc.On("UpdateItem", mock.Anything, int64(42), int64(999), mock.Anything).Return(nil, apperrors.ErrNotFound)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/items/999", strings.NewReader(`{"name":"X"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteHandler_Success(t *testing.T) {
	svc := new(mocks.MockItemService)
	router := setupRouter(svc)

	svc.On("DeleteItem", mock.Anything, int64(42), int64(1)).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/items/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Item deleted")
}

func TestDeleteHandler_NotFound(t *testing.T) {
	svc := new(mocks.MockItemService)
	router := setupRouter(svc)

	svc.On("DeleteItem", mock.Anything, int64(42), int64(999)).Return(apperrors.ErrNotFound)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/items/999", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandlers_NoAuth(t *testing.T) {
	svc := new(mocks.MockItemService)
	r := gin.New()
	r.GET("/items", ListHandler(svc))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/items", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
