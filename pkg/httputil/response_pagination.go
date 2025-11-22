package httputil

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// PaginationMeta contains pagination metadata
type PaginationMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
	TotalItems int64 `json:"total_items"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success    bool           `json:"success"`
	Data       interface{}    `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
	Timestamp  time.Time      `json:"timestamp"`
}

// RespondWithPagination sends a paginated JSON response
func RespondWithPagination(c *gin.Context, data interface{}, page, pageSize int, totalItems int64) {
	totalPages := int((totalItems + int64(pageSize) - 1) / int64(pageSize))

	c.JSON(http.StatusOK, PaginatedResponse{
		Success: true,
		Data:    data,
		Pagination: PaginationMeta{
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
			TotalItems: totalItems,
		},
		Timestamp: time.Now().UTC(),
	})
}
