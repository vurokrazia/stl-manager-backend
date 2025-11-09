package categories

import (
	"net/http"
	"testing"

	"stl-manager/internal/handlers/categories"
	"stl-manager/tests/integration/helpers"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestCreateCategory(t *testing.T) {
	tests := []struct {
		name     string
		body     interface{}
		wantCode int
	}{
		{
			name:     "create successfully",
			body:     categories.CreateCategoryRequest{Name: "test-create-" + uuid.New().String()[:8]},
			wantCode: http.StatusCreated,
		},
		{
			name:     "empty name fails",
			body:     categories.CreateCategoryRequest{Name: ""},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "invalid json fails",
			body:     "invalid",
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := helpers.POST("/categories", tt.body)
			resp := helpers.MakeRequest(t, req, handler.CreateCategory)

			assert.Equal(t, tt.wantCode, resp.Code)

			// Cleanup if created
			if resp.Code == http.StatusCreated {
				if id := resp.GetString("id"); id != "" {
					categoryID, _ := uuid.Parse(id)
					helpers.DeleteTestCategory(t, pgtype.UUID{Bytes: categoryID, Valid: true})
				}
			}
		})
	}
}
