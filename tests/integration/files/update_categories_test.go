package files

import (
	"net/http"
	"testing"

	"stl-manager/internal/handlers/files"
	"stl-manager/tests/integration/helpers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUpdateFileCategories(t *testing.T) {
	folder := helpers.CreateTestFolder(t, "test-folder")
	defer helpers.DeleteTestFolder(t, folder.ID)

	file := helpers.CreateTestFile(t, "test-file", "stl", folder.ID)
	defer helpers.DeleteTestFile(t, file.ID)

	cat1 := helpers.CreateTestCategory(t, "test-cat-1")
	defer helpers.DeleteTestCategory(t, cat1.ID)

	cat2 := helpers.CreateTestCategory(t, "test-cat-2")
	defer helpers.DeleteTestCategory(t, cat2.ID)

	tests := []struct {
		name     string
		fileID   string
		body     interface{}
		wantCode int
	}{
		{
			name:   "update categories successfully",
			fileID: uuid.UUID(file.ID.Bytes).String(),
			body: files.UpdateCategoriesRequest{
				CategoryIDs: []string{
					uuid.UUID(cat1.ID.Bytes).String(),
					uuid.UUID(cat2.ID.Bytes).String(),
				},
			},
			wantCode: http.StatusOK,
		},
		{
			name:   "update with empty categories",
			fileID: uuid.UUID(file.ID.Bytes).String(),
			body: files.UpdateCategoriesRequest{
				CategoryIDs: []string{},
			},
			wantCode: http.StatusOK,
		},
		{
			name:     "invalid file id",
			fileID:   "invalid",
			body:     files.UpdateCategoriesRequest{CategoryIDs: []string{}},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "file not found",
			fileID:   uuid.New().String(),
			body:     files.UpdateCategoriesRequest{CategoryIDs: []string{}},
			wantCode: http.StatusNotFound,
		},
		{
			name:     "invalid request body",
			fileID:   uuid.UUID(file.ID.Bytes).String(),
			body:     "invalid",
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := helpers.PATCH("/files/"+tt.fileID+"/categories", tt.body).WithURLParam("id", tt.fileID)
			resp := helpers.MakeRequest(t, req, handler.UpdateFileCategories)
			assert.Equal(t, tt.wantCode, resp.Code)

			if tt.wantCode == http.StatusOK {
				assert.NotNil(t, resp.Body["categories"])
				assert.NotNil(t, resp.Body["file_id"])
			}
		})
	}
}
