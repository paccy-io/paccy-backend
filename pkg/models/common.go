package models

type PaginationQuery struct {
	Page  int `query:"page" validate:"min=1"`
	Limit int `query:"limit" validate:"min=1,max=100"`
}

// PaginationQuery represents pagination query parameters
func DefaultPagination() PaginationQuery {
	return PaginationQuery{
		Page:  1,
		Limit: 10,
	}
}

// GetOffset calculates the offset for database queries
func (p PaginationQuery) GetOffset() int {
	return (p.Page - 1) * p.Limit
}

// GetSkip returns the skip value for MongoDB queries
func (p PaginationQuery) GetSkip() int64 {
	return int64((p.Page - 1) * p.Limit)
}

// GetLimit returns the limit value for MongoDB queries
func (p PaginationQuery) GetLimit() int64 {
	return int64(p.Limit)
}
