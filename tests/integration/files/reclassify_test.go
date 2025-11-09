package files

import (
	"net/http"
	"testing"

	"stl-manager/tests/integration/helpers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestReclassifyFile(t *testing.T) {
	folder := helpers.CreateTestFolder(t, "test-folder")
	defer helpers.DeleteTestFolder(t, folder.ID)

	file := helpers.CreateTestFile(t, "dragon-miniature", "stl", folder.ID)
	defer helpers.DeleteTestFile(t, file.ID)

	tests := []struct {
		name     string
		fileID   string
		wantCode int
	}{
		{
			name:     "reclassify requires OpenAI",
			fileID:   uuid.UUID(file.ID.Bytes).String(),
			wantCode: http.StatusServiceUnavailable, // OpenAI not enabled in tests
		},
		{
			name:     "invalid file id",
			fileID:   "invalid",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "file not found",
			fileID:   uuid.New().String(),
			wantCode: http.StatusNotFound, // File check happens before OpenAI check
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := helpers.POST("/files/"+tt.fileID+"/reclassify", nil).WithURLParam("id", tt.fileID)
			resp := helpers.MakeRequest(t, req, handler.ReclassifyFile)
			assert.Equal(t, tt.wantCode, resp.Code)
		})
	}
}
