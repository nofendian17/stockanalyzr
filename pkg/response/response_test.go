package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func TestResponseStructure(t *testing.T) {
	t.Run("Response struct should have correct fields", func(t *testing.T) {
		resp := Response{
			Status:  200,
			Success: true,
			Message: "OK",
			Data:    map[string]string{"key": "value"},
			Meta: &Meta{
				Page:       1,
				PerPage:    10,
				Total:      100,
				TotalPages: 10,
			},
		}

		assert.Equal(t, 200, resp.Status)
		assert.Equal(t, true, resp.Success)
		assert.Equal(t, "OK", resp.Message)
		assert.NotNil(t, resp.Data)
		assert.NotNil(t, resp.Meta)
		assert.Equal(t, 1, resp.Meta.Page)
		assert.Equal(t, 10, resp.Meta.PerPage)
		assert.Equal(t, int64(100), resp.Meta.Total)
	})
}

func TestPagination(t *testing.T) {
	t.Run("DefaultPagination should return correct defaults", func(t *testing.T) {
		p := DefaultPagination()
		assert.Equal(t, 1, p.Page)
		assert.Equal(t, 10, p.PerPage)
	})

	t.Run("Normalize should fix negative values", func(t *testing.T) {
		p := Pagination{Page: -1, PerPage: -5}
		p.Normalize()
		assert.Equal(t, 1, p.Page)
		assert.Equal(t, 10, p.PerPage)
	})

	t.Run("Normalize should cap per_page at 100", func(t *testing.T) {
		p := Pagination{Page: 1, PerPage: 200}
		p.Normalize()
		assert.Equal(t, 100, p.PerPage)
	})

	t.Run("Normalize should not change valid values", func(t *testing.T) {
		p := Pagination{Page: 5, PerPage: 25}
		p.Normalize()
		assert.Equal(t, 5, p.Page)
		assert.Equal(t, 25, p.PerPage)
	})

	t.Run("Offset should calculate correctly", func(t *testing.T) {
		p := Pagination{Page: 3, PerPage: 10}
		assert.Equal(t, 20, p.Offset())
	})

	t.Run("Limit should return per_page", func(t *testing.T) {
		p := Pagination{Page: 1, PerPage: 15}
		assert.Equal(t, 15, p.Limit())
	})
}

func TestNewMeta(t *testing.T) {
	t.Run("NewMeta should calculate total pages correctly", func(t *testing.T) {
		p := Pagination{Page: 1, PerPage: 10}
		meta := NewMeta(95, p)

		assert.Equal(t, 1, meta.Page)
		assert.Equal(t, 10, meta.PerPage)
		assert.Equal(t, int64(95), meta.Total)
		assert.Equal(t, 10, meta.TotalPages)
	})

	t.Run("NewMeta should handle exact division", func(t *testing.T) {
		p := Pagination{Page: 1, PerPage: 10}
		meta := NewMeta(100, p)

		assert.Equal(t, 10, meta.TotalPages)
	})

	t.Run("NewMeta should handle zero total", func(t *testing.T) {
		p := Pagination{Page: 1, PerPage: 10}
		meta := NewMeta(0, p)

		assert.Equal(t, 0, meta.TotalPages)
	})
}

func TestSuccess(t *testing.T) {
	t.Run("Success should return correct response", func(t *testing.T) {
		c, w := setupTestContext()
		data := map[string]string{"id": "123", "name": "Test"}

		Success(c, http.StatusOK, data)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Status)
		assert.True(t, resp.Success)
		assert.NotNil(t, resp.Data)
	})
}

func TestSuccessWithMessage(t *testing.T) {
	t.Run("SuccessWithMessage should include message", func(t *testing.T) {
		c, w := setupTestContext()
		data := map[string]string{"id": "123"}

		SuccessWithMessage(c, http.StatusCreated, "User created successfully", data)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.Status)
		assert.True(t, resp.Success)
		assert.Equal(t, "User created successfully", resp.Message)
	})
}

func TestSuccessWithPagination(t *testing.T) {
	t.Run("SuccessWithPagination should include meta", func(t *testing.T) {
		c, w := setupTestContext()
		data := []map[string]string{{"id": "1"}, {"id": "2"}}
		pagination := Pagination{Page: 2, PerPage: 10}

		SuccessWithPagination(c, http.StatusOK, data, 25, pagination)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.NotNil(t, resp.Meta)
		assert.Equal(t, 2, resp.Meta.Page)
		assert.Equal(t, 10, resp.Meta.PerPage)
		assert.Equal(t, int64(25), resp.Meta.Total)
		assert.Equal(t, 3, resp.Meta.TotalPages)
	})
}

func TestError(t *testing.T) {
	t.Run("Error should return correct error response", func(t *testing.T) {
		c, w := setupTestContext()

		Error(c, http.StatusBadRequest, "Something went wrong")

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.Status)
		assert.False(t, resp.Success)
		assert.Equal(t, "Something went wrong", resp.Message)
	})
}

func TestErrorWithDetails(t *testing.T) {
	t.Run("ErrorWithDetails should include errors", func(t *testing.T) {
		c, w := setupTestContext()
		errors := map[string]string{"email": "Invalid email format"}

		ErrorWithDetails(c, http.StatusUnprocessableEntity, "Validation failed", errors)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

		var resp Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Equal(t, "Validation failed", resp.Message)
		assert.NotNil(t, resp.Errors)
	})
}

func TestValidationError(t *testing.T) {
	t.Run("ValidationError should return 422", func(t *testing.T) {
		c, w := setupTestContext()
		errors := map[string]string{"name": "Name is required"}

		ValidationError(c, errors)

		assert.Equal(t, 422, w.Code)

		var resp Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, 422, resp.Status)
		assert.False(t, resp.Success)
		assert.Equal(t, "Validation failed", resp.Message)
	})
}

func TestNotFound(t *testing.T) {
	t.Run("NotFound should return 404 with resource name", func(t *testing.T) {
		c, w := setupTestContext()

		NotFound(c, "User")

		assert.Equal(t, http.StatusNotFound, w.Code)

		var resp Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.Status)
		assert.False(t, resp.Success)
		assert.Equal(t, "User not found", resp.Message)
	})
}

func TestUnauthorized(t *testing.T) {
	t.Run("Unauthorized with custom message", func(t *testing.T) {
		c, w := setupTestContext()

		Unauthorized(c, "Token expired")

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var resp Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "Token expired", resp.Message)
	})

	t.Run("Unauthorized with default message", func(t *testing.T) {
		c, w := setupTestContext()

		Unauthorized(c, "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var resp Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "Unauthorized", resp.Message)
	})
}

func TestForbidden(t *testing.T) {
	t.Run("Forbidden with custom message", func(t *testing.T) {
		c, w := setupTestContext()

		Forbidden(c, "Admin access required")

		assert.Equal(t, http.StatusForbidden, w.Code)

		var resp Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "Admin access required", resp.Message)
	})

	t.Run("Forbidden with default message", func(t *testing.T) {
		c, w := setupTestContext()

		Forbidden(c, "")

		assert.Equal(t, http.StatusForbidden, w.Code)

		var resp Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "Forbidden", resp.Message)
	})
}

func TestBadRequest(t *testing.T) {
	t.Run("BadRequest with custom message", func(t *testing.T) {
		c, w := setupTestContext()

		BadRequest(c, "Invalid JSON format")

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid JSON format", resp.Message)
	})

	t.Run("BadRequest with default message", func(t *testing.T) {
		c, w := setupTestContext()

		BadRequest(c, "")

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "Bad request", resp.Message)
	})
}

func TestInternalServerError(t *testing.T) {
	t.Run("InternalServerError with custom message", func(t *testing.T) {
		c, w := setupTestContext()

		InternalServerError(c, "Database connection failed")

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var resp Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "Database connection failed", resp.Message)
	})

	t.Run("InternalServerError with default message", func(t *testing.T) {
		c, w := setupTestContext()

		InternalServerError(c, "")

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var resp Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "Internal server error", resp.Message)
	})
}

func TestResponseJSONStructure(t *testing.T) {
	t.Run("Response JSON should omit empty fields", func(t *testing.T) {
		c, w := setupTestContext()

		Success(c, http.StatusOK, "data")

		var jsonMap map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &jsonMap)
		assert.NoError(t, err)

		// These should not be present when empty
		_, hasErrors := jsonMap["errors"]
		_, hasMeta := jsonMap["meta"]
		_, hasMessage := jsonMap["message"]

		assert.False(t, hasErrors, "errors should be omitted when nil")
		assert.False(t, hasMeta, "meta should be omitted when nil")
		assert.False(t, hasMessage, "message should be omitted when empty")
	})
}
