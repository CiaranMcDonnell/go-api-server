package interfaces

import (
	"context"

	"github.com/ciaranmcdonnell/go-api-server/internal/core/items/domain/models"
	sharedmodels "github.com/ciaranmcdonnell/go-api-server/models"
)

type ItemService interface {
	CreateItem(ctx context.Context, userID int64, dto models.CreateItemDTO) (*models.Item, error)
	GetItem(ctx context.Context, userID int64, itemID int64) (*models.Item, error)
	ListItems(ctx context.Context, userID int64, params sharedmodels.PaginationParams) ([]*models.Item, *sharedmodels.PaginationResult, error)
	UpdateItem(ctx context.Context, userID int64, itemID int64, dto models.UpdateItemDTO) (*models.Item, error)
	DeleteItem(ctx context.Context, userID int64, itemID int64) error
}
