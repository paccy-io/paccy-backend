package responses

import "time"

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Data    interface{}   `json:"data,omitempty"`
	Error   string        `json:"error,omitempty"`
	Meta    *MetaResponse `json:"meta,omitempty"`
}

// Meta contains pagination and additional metadata
type MetaResponse struct {
	Page       int `json:"page,omitempty"`
	Limit      int `json:"limit,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    int    `json:"code,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// HealthCheck represents health check response
type HealthCheck struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Uptime    string    `json:"uptime"`
	Database  string    `json:"database"`
	Redis     string    `json:"redis,omitempty"`
}

// SortOrder represents sort order for queries
type SortOrder struct {
	Field string `json:"field"`
	Order int    `json:"order"` // 1 for ascending, -1 for descending
}

// FilterQuery represents common filter parameters
type FilterQuery struct {
	Search   string     `query:"search"`
	Status   string     `query:"status"`
	SortBy   string     `query:"sort_by"`
	SortDesc bool       `query:"sort_desc"`
	FromDate *time.Time `query:"from_date"`
	ToDate   *time.Time `query:"to_date"`
}

// GetSortOrder returns MongoDB sort order
func (f FilterQuery) GetSortOrder() interface{} {
	if f.SortBy == "" {
		return map[string]int{"created_at": -1} // Default: newest first
	}

	order := 1
	if f.SortDesc {
		order = -1
	}

	return map[string]int{f.SortBy: order}
}
