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

// Helper functions for creating responses

// NewSuccessResponse creates a basic success response
func NewSuccessResponse() SuccessResponse {
	return SuccessResponse{
		Success: true,
	}
}

// NewSuccessResponseWithData creates a success response with data
func NewSuccessResponseWithData(data interface{}) SuccessResponse {
	return SuccessResponse{
		Success: true,
		Data:    data,
	}
}

// NewSuccessResponseWithMessage creates a success response with a message
func NewSuccessResponseWithMessage(message string) SuccessResponse {
	return SuccessResponse{
		Success: true,
		Message: message,
	}
}

// NewSuccessResponseWithDataAndMessage creates a success response with data and message
func NewSuccessResponseWithDataAndMessage(data interface{}, message string) SuccessResponse {
	return SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
	}
}

// NewSuccessResponseWithMeta creates a success response with data and metadata
func NewSuccessResponseWithMeta(data interface{}, meta interface{}) SuccessResponse {
	return SuccessResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	}
}

// NewPaginatedResponse creates a paginated response
func NewPaginatedResponse(data interface{}, total, perPage, currentPage int) PaginatedResponse {
	count := 0
	// Try to get count from data if it's a slice
	if dataSlice, ok := data.([]interface{}); ok {
		count = len(dataSlice)
	}

	totalPages := (total + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}

	hasMore := currentPage < totalPages

	return PaginatedResponse{
		Success: true,
		Data:    data,
		Meta: Pagination{
			Total:       total,
			Count:       count,
			PerPage:     perPage,
			CurrentPage: currentPage,
			TotalPages:  totalPages,
			HasMore:     hasMore,
		},
	}
}
