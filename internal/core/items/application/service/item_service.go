package service

import (
	"context"
	"fmt"

	"github.com/ciaranmcdonnell/go-api-server/internal/core/items/domain/interfaces"
	"github.com/ciaranmcdonnell/go-api-server/internal/core/items/domain/models"
)

type itemService struct {
	repo interfaces.ItemRepository
}

func NewItemService(repo interfaces.ItemRepository) interfaces.ItemService {
	return &itemService{repo: repo}
}

func (s *itemService) CreateItem(ctx context.Context, userID int64, dto models.CreateItemDTO) (*models.Item, error) {
	item := &models.Item{
		UserID:      userID,
		Name:        dto.Name,
		Description: dto.Description,
	}

	id, err := s.repo.Create(ctx, item)
	if err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, id)
}

func (s *itemService) GetItem(ctx context.Context, userID int64, itemID int64) (*models.Item, error) {
	item, err := s.repo.FindByID(ctx, itemID)
	if err != nil {
		return nil, err
	}

	if item.UserID != userID {
		return nil, fmt.Errorf("not found")
	}

	return item, nil
}

func (s *itemService) ListItems(ctx context.Context, userID int64, limit, offset int) ([]*models.Item, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.FindByFilter(ctx, models.ItemFilter{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
}

func (s *itemService) UpdateItem(ctx context.Context, userID int64, itemID int64, dto models.UpdateItemDTO) (*models.Item, error) {
	item, err := s.repo.FindByID(ctx, itemID)
	if err != nil {
		return nil, err
	}

	if item.UserID != userID {
		return nil, fmt.Errorf("not found")
	}

	if dto.Name != nil {
		item.Name = *dto.Name
	}
	if dto.Description != nil {
		item.Description = *dto.Description
	}

	if err := s.repo.Update(ctx, item); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, itemID)
}

func (s *itemService) DeleteItem(ctx context.Context, userID int64, itemID int64) error {
	item, err := s.repo.FindByID(ctx, itemID)
	if err != nil {
		return err
	}

	if item.UserID != userID {
		return fmt.Errorf("not found")
	}

	return s.repo.Delete(ctx, itemID)
}
