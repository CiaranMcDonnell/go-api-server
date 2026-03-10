package mocks

import (
	"context"

	"github.com/ciaranmcdonnell/go-api-server/internal/core/items/domain/models"
	"github.com/stretchr/testify/mock"
)

type MockItemRepository struct {
	mock.Mock
}

func (m *MockItemRepository) Create(ctx context.Context, item *models.Item) (int64, error) {
	args := m.Called(ctx, item)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockItemRepository) FindByID(ctx context.Context, id int64) (*models.Item, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Item), args.Error(1)
}

func (m *MockItemRepository) FindByFilter(ctx context.Context, filter models.ItemFilter) ([]*models.Item, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Item), args.Error(1)
}

func (m *MockItemRepository) Update(ctx context.Context, item *models.Item) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockItemRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
