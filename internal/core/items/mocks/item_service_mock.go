package mocks

import (
	"context"

	"github.com/ciaranmcdonnell/go-api-server/internal/core/items/domain/models"
	sharedmodels "github.com/ciaranmcdonnell/go-api-server/models"
	"github.com/stretchr/testify/mock"
)

type MockItemService struct {
	mock.Mock
}

func (m *MockItemService) CreateItem(ctx context.Context, userID int64, dto models.CreateItemDTO) (*models.Item, error) {
	args := m.Called(ctx, userID, dto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Item), args.Error(1)
}

func (m *MockItemService) GetItem(ctx context.Context, userID int64, itemID int64) (*models.Item, error) {
	args := m.Called(ctx, userID, itemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Item), args.Error(1)
}

func (m *MockItemService) ListItems(ctx context.Context, userID int64, params sharedmodels.PaginationParams) ([]*models.Item, *sharedmodels.PaginationResult, error) {
	args := m.Called(ctx, userID, params)
	var items []*models.Item
	if args.Get(0) != nil {
		items = args.Get(0).([]*models.Item)
	}
	var pagination *sharedmodels.PaginationResult
	if args.Get(1) != nil {
		pagination = args.Get(1).(*sharedmodels.PaginationResult)
	}
	return items, pagination, args.Error(2)
}

func (m *MockItemService) UpdateItem(ctx context.Context, userID int64, itemID int64, dto models.UpdateItemDTO) (*models.Item, error) {
	args := m.Called(ctx, userID, itemID, dto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Item), args.Error(1)
}

func (m *MockItemService) DeleteItem(ctx context.Context, userID int64, itemID int64) error {
	args := m.Called(ctx, userID, itemID)
	return args.Error(0)
}
