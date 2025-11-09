package categories

import (
	"net/http"
	"testing"

	"stl-manager/tests/integration/helpers"

	"github.com/stretchr/testify/assert"
)

func TestListCategories(t *testing.T) {
	// Create test categories
	cat1 := helpers.CreateTestCategory(t, "test-list")
	defer helpers.DeleteTestCategory(t, cat1.ID)

	cat2 := helpers.CreateTestCategory(t, "test-list")
	defer helpers.DeleteTestCategory(t, cat2.ID)

	tests := []struct {
		name string
		req  helpers.HTTPTestRequest
		want int
	}{
		{
			name: "list all categories",
			req:  helpers.GET("/categories"),
			want: http.StatusOK,
		},
		{
			name: "list with pagination",
			req:  helpers.GET("/categories").WithQueryParam("page", "1").WithQueryParam("page_size", "5"),
			want: http.StatusOK,
		},
		{
			name: "search categories",
			req:  helpers.GET("/categories").WithQueryParam("q", "test-list"),
			want: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := helpers.MakeRequest(t, tt.req, handler.ListCategories)
			assert.Equal(t, tt.want, resp.Code)
			helpers.AssertPaginatedResponse(t, resp)
		})
	}
}
