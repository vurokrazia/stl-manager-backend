package categories

import (
	"net/http"
	"testing"

	"stl-manager/tests/integration/helpers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSoftDeleteCategory(t *testing.T) {
	cat := helpers.CreateTestCategory(t, "test-delete")
	defer helpers.DeleteTestCategory(t, cat.ID)

	tests := []struct {
		name     string
		id       string
		wantCode int
	}{
		{
			name:     "delete successfully",
			id:       uuid.UUID(cat.ID.Bytes).String(),
			wantCode: http.StatusOK,
		},
		{
			name:     "invalid id",
			id:       "invalid",
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := helpers.DELETE("/categories/"+tt.id).WithURLParam("id", tt.id)
			resp := helpers.MakeRequest(t, req, handler.SoftDeleteCategory)
			assert.Equal(t, tt.wantCode, resp.Code)
		})
	}
}

func TestSoftDeleteHidesCategory(t *testing.T) {
	cat := helpers.CreateTestCategory(t, "test-hidden")
	defer helpers.DeleteTestCategory(t, cat.ID)

	// Verify visible before delete
	req1 := helpers.GET("/categories").WithQueryParam("q", "test-hidden")
	resp1 := helpers.MakeRequest(t, req1, handler.ListCategories)
	items1 := resp1.GetArray("items")
	assert.GreaterOrEqual(t, len(items1), 1)

	// Soft delete
	helpers.SoftDeleteTestCategory(t, cat.ID)

	// Verify hidden after delete
	req2 := helpers.GET("/categories").WithQueryParam("q", "test-hidden")
	resp2 := helpers.MakeRequest(t, req2, handler.ListCategories)
	items2 := resp2.GetArray("items")
	assert.Less(t, len(items2), len(items1))
}
