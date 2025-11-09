package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// AssertSuccessResponse asserts a successful HTTP response
func AssertSuccessResponse(t *testing.T, resp *HTTPTestResponse, expectedCode int) {
	assert.Equal(t, expectedCode, resp.Code, "Expected status code %d, got %d", expectedCode, resp.Code)
}

// AssertErrorResponse asserts an error HTTP response
func AssertErrorResponse(t *testing.T, resp *HTTPTestResponse, expectedCode int) {
	assert.Equal(t, expectedCode, resp.Code)
	assert.Contains(t, resp.Body, "error", "Response should contain 'error' field")
}

// AssertHasFields asserts that response body has required fields
func AssertHasFields(t *testing.T, body map[string]interface{}, fields ...string) {
	for _, field := range fields {
		assert.Contains(t, body, field, "Response should contain field: %s", field)
	}
}

// AssertPaginatedResponse asserts a paginated response structure
func AssertPaginatedResponse(t *testing.T, resp *HTTPTestResponse) {
	AssertHasFields(t, resp.Body, "items", "total", "page", "page_size", "total_pages")
}
