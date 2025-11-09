package files

import (
	"net/http"
	"testing"

	"stl-manager/tests/integration/helpers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetFile(t *testing.T) {
	folder := helpers.CreateTestFolder(t, "test-folder")
	defer helpers.DeleteTestFolder(t, folder.ID)

	file := helpers.CreateTestFile(t, "test-file", "stl", folder.ID)
	defer helpers.DeleteTestFile(t, file.ID)

	tests := []struct {
		name     string
		id       string
		wantCode int
	}{
		{
			name:     "get existing file",
			id:       uuid.UUID(file.ID.Bytes).String(),
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
			req := helpers.GET("/files/" + tt.id).WithURLParam("id", tt.id)
			resp := helpers.MakeRequest(t, req, handler.GetFile)
			assert.Equal(t, tt.wantCode, resp.Code)

			if tt.wantCode == http.StatusOK {
				assert.NotNil(t, resp.Body["id"])
				assert.NotNil(t, resp.Body["file_name"])
				assert.NotNil(t, resp.Body["categories"])
			}
		})
	}
}
