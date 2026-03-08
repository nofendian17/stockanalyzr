// Package response provides standardized HTTP response structures for REST APIs.
// It includes support for success responses, error responses, and pagination metadata.
package response

import (
	"github.com/gin-gonic/gin"
)

// Response is the standard API response wrapper.
type Response struct {
	Status  int         `json:"status"`            // HTTP status code
	Success bool        `json:"success"`           // Indicates if the request was successful
	Message string      `json:"message,omitempty"` // Human-readable message
	Data    interface{} `json:"data,omitempty"`    // Response payload
	Meta    *Meta       `json:"meta,omitempty"`    // Metadata (pagination, etc.)
	Errors  interface{} `json:"errors,omitempty"`  // Error details
}

// Meta contains metadata for the response, primarily used for pagination.
type Meta struct {
	Page       int   `json:"page,omitempty"`        // Current page number
	PerPage    int   `json:"per_page,omitempty"`    // Items per page
	Total      int64 `json:"total,omitempty"`       // Total number of items
	TotalPages int   `json:"total_pages,omitempty"` // Total number of pages
}

// Pagination holds pagination parameters from the request.
type Pagination struct {
	Page    int `json:"page" form:"page"`         // Page number (1-based)
	PerPage int `json:"per_page" form:"per_page"` // Items per page
}

// DefaultPagination returns a Pagination with default values.
func DefaultPagination() Pagination {
	return Pagination{
		Page:    1,
		PerPage: 10,
	}
}

// Normalize ensures pagination values are valid.
func (p *Pagination) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 {
		p.PerPage = 10
	}
	if p.PerPage > 100 {
		p.PerPage = 100
	}
}

// Offset calculates the SQL offset based on page and per_page.
func (p *Pagination) Offset() int {
	return (p.Page - 1) * p.PerPage
}

// Limit returns the limit value for SQL queries.
func (p *Pagination) Limit() int {
	return p.PerPage
}

// NewMeta creates pagination metadata from total count and pagination params.
func NewMeta(total int64, p Pagination) *Meta {
	totalPages := int(total) / p.PerPage
	if int(total)%p.PerPage > 0 {
		totalPages++
	}
	return &Meta{
		Page:       p.Page,
		PerPage:    p.PerPage,
		Total:      total,
		TotalPages: totalPages,
	}
}

// Success sends a successful response with data.
func Success(c *gin.Context, status int, data interface{}) {
	c.JSON(status, Response{
		Status:  status,
		Success: true,
		Data:    data,
	})
}

// SuccessWithMessage sends a successful response with a custom message.
func SuccessWithMessage(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, Response{
		Status:  status,
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SuccessWithPagination sends a successful response with paginated data.
func SuccessWithPagination(c *gin.Context, status int, data interface{}, total int64, p Pagination) {
	c.JSON(status, Response{
		Status:  status,
		Success: true,
		Data:    data,
		Meta:    NewMeta(total, p),
	})
}

// Error sends an error response.
func Error(c *gin.Context, status int, message string) {
	c.JSON(status, Response{
		Status:  status,
		Success: false,
		Message: message,
	})
}

// ErrorWithDetails sends an error response with detailed errors.
func ErrorWithDetails(c *gin.Context, status int, message string, errors interface{}) {
	c.JSON(status, Response{
		Status:  status,
		Success: false,
		Message: message,
		Errors:  errors,
	})
}

// ValidationError sends a 422 Unprocessable Entity response for validation errors.
func ValidationError(c *gin.Context, errors interface{}) {
	c.JSON(422, Response{
		Status:  422,
		Success: false,
		Message: "Validation failed",
		Errors:  errors,
	})
}

// NotFound sends a 404 Not Found response.
func NotFound(c *gin.Context, resource string) {
	c.JSON(404, Response{
		Status:  404,
		Success: false,
		Message: resource + " not found",
	})
}

// Unauthorized sends a 401 Unauthorized response.
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "Unauthorized"
	}
	c.JSON(401, Response{
		Status:  401,
		Success: false,
		Message: message,
	})
}

// Forbidden sends a 403 Forbidden response.
func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = "Forbidden"
	}
	c.JSON(403, Response{
		Status:  403,
		Success: false,
		Message: message,
	})
}

// BadRequest sends a 400 Bad Request response.
func BadRequest(c *gin.Context, message string) {
	if message == "" {
		message = "Bad request"
	}
	c.JSON(400, Response{
		Status:  400,
		Success: false,
		Message: message,
	})
}

// InternalServerError sends a 500 Internal Server Error response.
func InternalServerError(c *gin.Context, message string) {
	if message == "" {
		message = "Internal server error"
	}
	c.JSON(500, Response{
		Status:  500,
		Success: false,
		Message: message,
	})
}
