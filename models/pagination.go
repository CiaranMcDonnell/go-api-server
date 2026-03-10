package models

type PaginationParams struct {
	Page    int `form:"page"`
	PerPage int `form:"per_page"`
}

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
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	return PaginationParams{
		Page:    page,
		PerPage: perPage,
	}
}

func (p PaginationParams) Offset() int {
	return (p.Page - 1) * p.PerPage
}

// FetchLimit returns PerPage+1 so callers can detect whether a next page exists
// without running a separate COUNT query.
func (p PaginationParams) FetchLimit() int {
	return p.PerPage + 1
}

func NewPaginationResult(params PaginationParams, fetched int) PaginationResult {
	return PaginationResult{
		Page:        params.Page,
		PerPage:     params.PerPage,
		HasNextPage: fetched > params.PerPage,
		HasPrevPage: params.Page > 1,
	}
}
