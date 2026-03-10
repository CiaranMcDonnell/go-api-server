package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ciaranmcdonnell/go-api-server/internal/core/items/domain/models"
	"github.com/ciaranmcdonnell/go-api-server/internal/core/items/mocks"
	"github.com/ciaranmcdonnell/go-api-server/pkg/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestItem(id, userID int64) *models.Item {
	return &models.Item{
		ID:          id,
		UserID:      userID,
		Name:        "Test Item",
		Description: "A test item",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func TestCreateItem_Success(t *testing.T) {
	repo := new(mocks.MockItemRepository)
	svc := NewItemService(repo)
	ctx := context.Background()

	dto := models.CreateItemDTO{Name: "New Item", Description: "desc"}
	created := newTestItem(1, 42)

	repo.On("Create", ctx, mock.AnythingOfType("*models.Item")).Return(int64(1), nil)
	repo.On("FindByID", ctx, int64(1)).Return(created, nil)

	item, err := svc.CreateItem(ctx, 42, dto)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), item.ID)
	assert.Equal(t, int64(42), item.UserID)
	repo.AssertExpectations(t)
}

func TestCreateItem_RepoError(t *testing.T) {
	repo := new(mocks.MockItemRepository)
	svc := NewItemService(repo)
	ctx := context.Background()

	dto := models.CreateItemDTO{Name: "New Item"}

	repo.On("Create", ctx, mock.Anything).Return(int64(0), fmt.Errorf("db error"))

	item, err := svc.CreateItem(ctx, 42, dto)

	assert.Error(t, err)
	assert.Nil(t, item)
	repo.AssertExpectations(t)
}

func TestGetItem_Success(t *testing.T) {
	repo := new(mocks.MockItemRepository)
	svc := NewItemService(repo)
	ctx := context.Background()

	expected := newTestItem(1, 42)
	repo.On("FindByID", ctx, int64(1)).Return(expected, nil)

	item, err := svc.GetItem(ctx, 42, 1)

	assert.NoError(t, err)
	assert.Equal(t, expected, item)
	repo.AssertExpectations(t)
}

func TestGetItem_NotFound(t *testing.T) {
	repo := new(mocks.MockItemRepository)
	svc := NewItemService(repo)
	ctx := context.Background()

	repo.On("FindByID", ctx, int64(999)).Return(nil, fmt.Errorf("not found"))

	item, err := svc.GetItem(ctx, 42, 999)

	assert.Error(t, err)
	assert.Nil(t, item)
	repo.AssertExpectations(t)
}

func TestGetItem_WrongOwner(t *testing.T) {
	repo := new(mocks.MockItemRepository)
	svc := NewItemService(repo)
	ctx := context.Background()

	item := newTestItem(1, 42)
	repo.On("FindByID", ctx, int64(1)).Return(item, nil)

	result, err := svc.GetItem(ctx, 99, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, apperrors.ErrNotFound))
	repo.AssertExpectations(t)
}

func TestListItems_Success(t *testing.T) {
	repo := new(mocks.MockItemRepository)
	svc := NewItemService(repo)
	ctx := context.Background()

	expected := []*models.Item{newTestItem(1, 42), newTestItem(2, 42)}
	repo.On("FindByFilter", ctx, models.ItemFilter{UserID: 42, Limit: 20, Offset: 0}).Return(expected, nil)

	items, err := svc.ListItems(ctx, 42, 20, 0)

	assert.NoError(t, err)
	assert.Len(t, items, 2)
	repo.AssertExpectations(t)
}

func TestListItems_ClampsLimit(t *testing.T) {
	repo := new(mocks.MockItemRepository)
	svc := NewItemService(repo)
	ctx := context.Background()

	repo.On("FindByFilter", ctx, models.ItemFilter{UserID: 42, Limit: 20, Offset: 0}).Return([]*models.Item{}, nil)

	_, err := svc.ListItems(ctx, 42, 999, 0)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestListItems_NegativeOffset(t *testing.T) {
	repo := new(mocks.MockItemRepository)
	svc := NewItemService(repo)
	ctx := context.Background()

	repo.On("FindByFilter", ctx, models.ItemFilter{UserID: 42, Limit: 20, Offset: 0}).Return([]*models.Item{}, nil)

	_, err := svc.ListItems(ctx, 42, 20, -5)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestUpdateItem_Success(t *testing.T) {
	repo := new(mocks.MockItemRepository)
	svc := NewItemService(repo)
	ctx := context.Background()

	existing := newTestItem(1, 42)
	newName := "Updated Name"
	updated := newTestItem(1, 42)
	updated.Name = newName

	repo.On("FindByID", ctx, int64(1)).Return(existing, nil).Once()
	repo.On("Update", ctx, mock.AnythingOfType("*models.Item")).Return(nil)
	repo.On("FindByID", ctx, int64(1)).Return(updated, nil).Once()

	dto := models.UpdateItemDTO{Name: &newName}
	item, err := svc.UpdateItem(ctx, 42, 1, dto)

	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", item.Name)
	repo.AssertExpectations(t)
}

func TestUpdateItem_WrongOwner(t *testing.T) {
	repo := new(mocks.MockItemRepository)
	svc := NewItemService(repo)
	ctx := context.Background()

	existing := newTestItem(1, 42)
	repo.On("FindByID", ctx, int64(1)).Return(existing, nil)

	newName := "Hacked"
	dto := models.UpdateItemDTO{Name: &newName}
	item, err := svc.UpdateItem(ctx, 99, 1, dto)

	assert.Error(t, err)
	assert.Nil(t, item)
	assert.True(t, errors.Is(err, apperrors.ErrNotFound))
	repo.AssertExpectations(t)
}

func TestUpdateItem_PartialUpdate(t *testing.T) {
	repo := new(mocks.MockItemRepository)
	svc := NewItemService(repo)
	ctx := context.Background()

	existing := newTestItem(1, 42)
	existing.Name = "Original"
	existing.Description = "Original desc"

	newDesc := "Updated desc"
	updated := newTestItem(1, 42)
	updated.Name = "Original"
	updated.Description = newDesc

	repo.On("FindByID", ctx, int64(1)).Return(existing, nil).Once()
	repo.On("Update", ctx, mock.MatchedBy(func(item *models.Item) bool {
		return item.Name == "Original" && item.Description == "Updated desc"
	})).Return(nil)
	repo.On("FindByID", ctx, int64(1)).Return(updated, nil).Once()

	dto := models.UpdateItemDTO{Description: &newDesc}
	item, err := svc.UpdateItem(ctx, 42, 1, dto)

	assert.NoError(t, err)
	assert.Equal(t, "Original", item.Name)
	assert.Equal(t, "Updated desc", item.Description)
	repo.AssertExpectations(t)
}

func TestDeleteItem_Success(t *testing.T) {
	repo := new(mocks.MockItemRepository)
	svc := NewItemService(repo)
	ctx := context.Background()

	existing := newTestItem(1, 42)
	repo.On("FindByID", ctx, int64(1)).Return(existing, nil)
	repo.On("Delete", ctx, int64(1)).Return(nil)

	err := svc.DeleteItem(ctx, 42, 1)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDeleteItem_WrongOwner(t *testing.T) {
	repo := new(mocks.MockItemRepository)
	svc := NewItemService(repo)
	ctx := context.Background()

	existing := newTestItem(1, 42)
	repo.On("FindByID", ctx, int64(1)).Return(existing, nil)

	err := svc.DeleteItem(ctx, 99, 1)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrNotFound))
	repo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestDeleteItem_NotFound(t *testing.T) {
	repo := new(mocks.MockItemRepository)
	svc := NewItemService(repo)
	ctx := context.Background()

	repo.On("FindByID", ctx, int64(999)).Return(nil, fmt.Errorf("not found"))

	err := svc.DeleteItem(ctx, 42, 999)

	assert.Error(t, err)
	repo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}
