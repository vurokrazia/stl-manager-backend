package categories

import (
	"net/http"
	"testing"

	"stl-manager/internal/handlers/categories"
	"stl-manager/tests/integration/helpers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUpdateCategory(t *testing.T) {
	cat := helpers.CreateTestCategory(t, "test-update")
	defer helpers.DeleteTestCategory(t, cat.ID)

	tests := []struct {
		name     string
		id       string
		body     interface{}
		wantCode int
	}{
		{
			name:     "update successfully",
			id:       uuid.UUID(cat.ID.Bytes).String(),
			body:     categories.UpdateCategoryRequest{Name: "updated-" + uuid.New().String()[:8]},
			wantCode: http.StatusOK,
		},
		{
			name:     "empty name fails",
			id:       uuid.UUID(cat.ID.Bytes).String(),
			body:     categories.UpdateCategoryRequest{Name: ""},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "invalid id",
			id:       "invalid",
			body:     categories.UpdateCategoryRequest{Name: "test"},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := helpers.PUT("/categories/"+tt.id, tt.body).WithURLParam("id", tt.id)
			resp := helpers.MakeRequest(t, req, handler.UpdateCategory)
			assert.Equal(t, tt.wantCode, resp.Code)
		})
	}
}
