// internal/handlers/response.go
package handlers

// SuccessResponse represents a successful API response
// @Description Standard success response wrapper
type SuccessResponse struct {
	Success bool        `json:"success" example:"true"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty" example:"Operation completed successfully"`
	Meta    interface{} `json:"meta,omitempty"`
}

// PaginatedResponse represents a paginated API response
// @Description Paginated response wrapper with metadata
type PaginatedResponse struct {
	Success bool        `json:"success" example:"true"`
	Data    interface{} `json:"data"`
	Meta    Pagination  `json:"meta"`
}

// Pagination contains pagination metadata
// @Description Pagination metadata for list responses
type Pagination struct {
	Total       int  `json:"total" example:"100"`
	Count       int  `json:"count" example:"20"`
	PerPage     int  `json:"per_page" example:"20"`
	CurrentPage int  `json:"current_page" example:"1"`
	TotalPages  int  `json:"total_pages" example:"5"`
	HasMore     bool `json:"has_more" example:"true"`
}

// Note: Helper functions for creating responses are in helpers.go
// Use RespondSuccess, RespondSuccessWithMessage, RespondCreated, etc.
