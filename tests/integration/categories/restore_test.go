package categories

import (
	"net/http"
	"testing"

	"stl-manager/tests/integration/helpers"

	"github.com/stretchr/testify/assert"
)

func TestRestoreCategory(t *testing.T) {
	cat := helpers.CreateTestCategory(t, "test-restore")
	defer helpers.DeleteTestCategory(t, cat.ID)

	// Soft delete first
	helpers.SoftDeleteTestCategory(t, cat.ID)

	tests := []struct {
		name     string
		id       string
		wantCode int
	}{
		{
			name:     "restore successfully",
			id:       cat.ID,
			wantCode: http.StatusOK,
		},
		{
			name:     "invalid id still succeeds (idempotent)",
			id:       "invalid",
			wantCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := helpers.POST("/categories/"+tt.id+"/restore", nil).WithURLParam("id", tt.id)
			resp := helpers.MakeRequest(t, req, handler.RestoreCategory)
			assert.Equal(t, tt.wantCode, resp.Code)
		})
	}
}
