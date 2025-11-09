package categories

import (
	"net/http"
	"testing"

	"stl-manager/tests/integration/helpers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetCategory(t *testing.T) {
	cat := helpers.CreateTestCategory(t, "test-get")
	defer helpers.DeleteTestCategory(t, cat.ID)

	tests := []struct {
		name     string
		id       string
		wantCode int
	}{
		{
			name:     "get existing category",
			id:       uuid.UUID(cat.ID.Bytes).String(),
			wantCode: http.StatusOK,
		},
		{
			name:     "invalid id",
			id:       "invalid",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "not found",
			id:       uuid.New().String(),
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := helpers.GET("/categories/" + tt.id).WithURLParam("id", tt.id)
			resp := helpers.MakeRequest(t, req, handler.GetCategory)
			assert.Equal(t, tt.wantCode, resp.Code)
		})
	}
}
