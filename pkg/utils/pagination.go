package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type PaginationParams struct {
	Page   int
	Limit  int
	Offset int
}

// GetPagination extracts page and limit from query params
// Defaults: page=1, limit=20
func GetPagination(c *gin.Context) PaginationParams {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	return PaginationParams{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
}

type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	TotalItems int64       `json:"total_items"`
	TotalPages int         `json:"total_pages"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
}

// BuildPaginatedResponse wraps data with pagination metadata
func BuildPaginatedResponse(items interface{}, total int64, params PaginationParams) PaginatedResponse {
	totalPages := int(total) / params.Limit
	if int(total)%params.Limit != 0 {
		totalPages++
	}
	return PaginatedResponse{
		Items:      items,
		TotalItems: total,
		TotalPages: totalPages,
		Page:       params.Page,
		Limit:      params.Limit,
	}
}
