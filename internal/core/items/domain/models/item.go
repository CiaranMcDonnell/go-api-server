package models

import (
	"time"

	sharedmodels "github.com/ciaranmcdonnell/go-api-server/models"
)

type Item struct {
	ID          int64
	UserID      int64
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateItemDTO struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"max=2000"`
}

type UpdateItemDTO struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type ItemResponse struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ItemFilter struct {
	UserID     int64
	Pagination sharedmodels.PaginationParams
}

func (i *Item) ToResponse() *ItemResponse {
	return &ItemResponse{
		ID:          i.ID,
		UserID:      i.UserID,
		Name:        i.Name,
		Description: i.Description,
		CreatedAt:   i.CreatedAt,
		UpdatedAt:   i.UpdatedAt,
	}
}

func ItemsToResponses(items []*Item) []*ItemResponse {
	responses := make([]*ItemResponse, len(items))
	for idx, item := range items {
		responses[idx] = item.ToResponse()
	}
	return responses
}
