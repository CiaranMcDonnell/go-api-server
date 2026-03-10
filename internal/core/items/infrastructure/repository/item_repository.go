package repository

import (
	"context"
	"fmt"

	"github.com/ciaranmcdonnell/go-api-server/internal/core/items/domain/interfaces"
	"github.com/ciaranmcdonnell/go-api-server/internal/core/items/domain/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type itemRepository struct {
	db *pgxpool.Pool
}

func NewItemRepository(db *pgxpool.Pool) interfaces.ItemRepository {
	return &itemRepository{db: db}
}

func (r *itemRepository) Create(ctx context.Context, item *models.Item) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx,
		`INSERT INTO items (user_id, name, description) VALUES ($1, $2, $3) RETURNING id`,
		item.UserID, item.Name, item.Description,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("creating item: %w", err)
	}
	return id, nil
}

func (r *itemRepository) FindByID(ctx context.Context, id int64) (*models.Item, error) {
	var item models.Item
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, name, description, created_at, updated_at FROM items WHERE id = $1`,
		id,
	).Scan(&item.ID, &item.UserID, &item.Name, &item.Description, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("finding item: %w", err)
	}
	return &item, nil
}

func (r *itemRepository) FindByFilter(ctx context.Context, filter models.ItemFilter) ([]*models.Item, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, name, description, created_at, updated_at
		 FROM items WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		filter.UserID, filter.Limit, filter.Offset,
	)
	if err != nil {
		return nil, fmt.Errorf("listing items: %w", err)
	}
	defer rows.Close()

	var items []*models.Item
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.ID, &item.UserID, &item.Name, &item.Description, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning item: %w", err)
		}
		items = append(items, &item)
	}
	return items, rows.Err()
}

func (r *itemRepository) Update(ctx context.Context, item *models.Item) error {
	_, err := r.db.Exec(ctx,
		`UPDATE items SET name = $1, description = $2, updated_at = NOW() WHERE id = $3`,
		item.Name, item.Description, item.ID,
	)
	if err != nil {
		return fmt.Errorf("updating item: %w", err)
	}
	return nil
}

func (r *itemRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM items WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting item: %w", err)
	}
	return nil
}
