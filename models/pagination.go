package models

// PaginationParams defines the parameters for pagination
type PaginationParams struct {
	Page    int `json:"page" form:"page"`
	PerPage int `json:"per_page" form:"per_page"`
}

// PaginationResult contains the pagination metadata
type PaginationResult struct {
	Page        int  `json:"page"`
	PerPage     int  `json:"per_page"`
	HasNextPage bool `json:"has_next_page"`
	HasPrevPage bool `json:"has_prev_page"`
}

func NewPaginationParams(page, perPage int) PaginationParams {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100 // Maximum per page
	}

	return PaginationParams{
		Page:    page,
		PerPage: perPage,
	}
}

func (p PaginationParams) GetOffset() int {
	return (p.Page - 1) * p.PerPage
}

func (p PaginationParams) GetLimit() int {
	return p.PerPage
}

func NewPaginationResult(params PaginationParams, hasNextPage bool) PaginationResult {
	return PaginationResult{
		Page:        params.Page,
		PerPage:     params.PerPage,
		HasNextPage: hasNextPage,
		HasPrevPage: params.Page > 1,
	}
}
