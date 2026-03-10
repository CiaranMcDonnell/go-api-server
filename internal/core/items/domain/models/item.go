package models

import "time"

// Domain model — no JSON tags, used internally by services and repositories.
type Item struct {
	ID          int64
	UserID      int64
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// API input for creating an item.
type CreateItemDTO struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"max=2000"`
}

// API input for updating an item.
type UpdateItemDTO struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// API output — safe to serialize.
type ItemResponse struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ItemFilter struct {
	UserID int64
	Limit  int
	Offset int
}

// ToResponse converts a domain Item to an API response.
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

// ItemsToResponses converts a slice of domain Items to API responses.
func ItemsToResponses(items []*Item) []*ItemResponse {
	responses := make([]*ItemResponse, len(items))
	for idx, item := range items {
		responses[idx] = item.ToResponse()
	}
	return responses
}
