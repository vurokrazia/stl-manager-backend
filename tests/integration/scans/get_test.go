package scans

import (
	"net/http"
	"testing"

	"stl-manager/tests/integration/helpers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetScan(t *testing.T) {
	scan := helpers.CreateTestScan(t, "completed")
	defer helpers.DeleteTestScan(t, scan.ID)

	tests := []struct {
		name     string
		id       string
		wantCode int
	}{
		{
			name:     "get existing scan",
			id:       uuid.UUID(scan.ID.Bytes).String(),
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
			req := helpers.GET("/scans/" + tt.id).WithURLParam("id", tt.id)
			resp := helpers.MakeRequest(t, req, handler.GetScan)
			assert.Equal(t, tt.wantCode, resp.Code)

			if tt.wantCode == http.StatusOK {
				assert.NotNil(t, resp.Body["id"])
				assert.NotNil(t, resp.Body["status"])
				assert.Equal(t, "completed", resp.Body["status"])
				assert.NotNil(t, resp.Body["found"])
				assert.NotNil(t, resp.Body["processed"])
				assert.NotNil(t, resp.Body["progress"])
				assert.NotNil(t, resp.Body["created_at"])
				assert.NotNil(t, resp.Body["updated_at"])
			}
		})
	}
}
