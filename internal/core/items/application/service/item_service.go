package service

import (
	"context"

	"github.com/ciaranmcdonnell/go-api-server/internal/core/items/domain/interfaces"
	"github.com/ciaranmcdonnell/go-api-server/internal/core/items/domain/models"
	"github.com/ciaranmcdonnell/go-api-server/internal/database"
	sharedmodels "github.com/ciaranmcdonnell/go-api-server/models"
	"github.com/ciaranmcdonnell/go-api-server/pkg/apperrors"
)

type itemService struct {
	repo  interfaces.ItemRepository
	txMgr database.TxManager
}

func NewItemService(repo interfaces.ItemRepository, txMgr database.TxManager) interfaces.ItemService {
	return &itemService{repo: repo, txMgr: txMgr}
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
		return nil, apperrors.ErrNotFound
	}

	return item, nil
}

func (s *itemService) ListItems(ctx context.Context, userID int64, params sharedmodels.PaginationParams) ([]*models.Item, *sharedmodels.PaginationResult, error) {
	items, err := s.repo.FindByFilter(ctx, models.ItemFilter{
		UserID:     userID,
		Pagination: params,
	})
	if err != nil {
		return nil, nil, err
	}

	pagination := sharedmodels.NewPaginationResult(params, len(items))

	if len(items) > params.PerPage {
		items = items[:params.PerPage]
	}

	return items, &pagination, nil
}

func (s *itemService) UpdateItem(ctx context.Context, userID int64, itemID int64, dto models.UpdateItemDTO) (*models.Item, error) {
	var result *models.Item
	err := s.txMgr.WithTx(ctx, func(ctx context.Context) error {
		item, err := s.repo.FindByID(ctx, itemID)
		if err != nil {
			return err
		}

		if item.UserID != userID {
			return apperrors.ErrNotFound
		}

		if dto.Name != nil {
			item.Name = *dto.Name
		}
		if dto.Description != nil {
			item.Description = *dto.Description
		}

		if err := s.repo.Update(ctx, item); err != nil {
			return err
		}

		result, err = s.repo.FindByID(ctx, itemID)
		return err
	})
	return result, err
}

func (s *itemService) DeleteItem(ctx context.Context, userID int64, itemID int64) error {
	return s.txMgr.WithTx(ctx, func(ctx context.Context) error {
		item, err := s.repo.FindByID(ctx, itemID)
		if err != nil {
			return err
		}

		if item.UserID != userID {
			return apperrors.ErrNotFound
		}

		return s.repo.Delete(ctx, itemID)
	})
}
