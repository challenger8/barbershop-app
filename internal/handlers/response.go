// internal/handlers/response.go
package handlers

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Meta    Pagination  `json:"meta"`
}

// Pagination contains pagination metadata
type Pagination struct {
	Total       int  `json:"total"`
	Count       int  `json:"count"`
	PerPage     int  `json:"per_page"`
	CurrentPage int  `json:"current_page"`
	TotalPages  int  `json:"total_pages"`
	HasMore     bool `json:"has_more"`
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
