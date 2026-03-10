package interfaces

import (
	"context"

	"github.com/ciaranmcdonnell/go-api-server/internal/core/items/domain/models"
)

type ItemRepository interface {
	Create(ctx context.Context, item *models.Item) (int64, error)
	FindByID(ctx context.Context, id int64) (*models.Item, error)
	FindByFilter(ctx context.Context, filter models.ItemFilter) ([]*models.Item, error)
	Update(ctx context.Context, item *models.Item) error
	Delete(ctx context.Context, id int64) error
}
