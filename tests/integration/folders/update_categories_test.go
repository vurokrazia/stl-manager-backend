package folders

import (
	"net/http"
	"testing"

	"stl-manager/tests/integration/helpers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUpdateFolderCategories(t *testing.T) {
	folder := helpers.CreateTestFolder(t, "test-folder")
	defer helpers.DeleteTestFolder(t, folder.ID)

	cat1 := helpers.CreateTestCategory(t, "test-cat-1")
	defer helpers.DeleteTestCategory(t, cat1.ID)

	cat2 := helpers.CreateTestCategory(t, "test-cat-2")
	defer helpers.DeleteTestCategory(t, cat2.ID)

	tests := []struct {
		name     string
		folderID string
		body     interface{}
		wantCode int
	}{
		{
			name:     "update categories successfully",
			folderID: uuid.UUID(folder.ID.Bytes).String(),
			body: map[string]interface{}{
				"category_ids": []string{
					uuid.UUID(cat1.ID.Bytes).String(),
					uuid.UUID(cat2.ID.Bytes).String(),
				},
				"apply_to_stl":        false,
				"apply_to_zip":        false,
				"apply_to_rar":        false,
				"apply_to_subfolders": false,
			},
			wantCode: http.StatusOK,
		},
		{
			name:     "update with empty categories",
			folderID: uuid.UUID(folder.ID.Bytes).String(),
			body: map[string]interface{}{
				"category_ids":        []string{},
				"apply_to_stl":        false,
				"apply_to_zip":        false,
				"apply_to_rar":        false,
				"apply_to_subfolders": false,
			},
			wantCode: http.StatusOK,
		},
		{
			name:     "apply to stl files",
			folderID: uuid.UUID(folder.ID.Bytes).String(),
			body: map[string]interface{}{
				"category_ids": []string{
					uuid.UUID(cat1.ID.Bytes).String(),
				},
				"apply_to_stl":        true,
				"apply_to_zip":        false,
				"apply_to_rar":        false,
				"apply_to_subfolders": false,
			},
			wantCode: http.StatusOK,
		},
		{
			name:     "invalid folder id",
			folderID: "invalid",
			body: map[string]interface{}{
				"category_ids": []string{},
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "invalid request body",
			folderID: uuid.UUID(folder.ID.Bytes).String(),
			body:     "invalid",
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := helpers.PATCH("/folders/"+tt.folderID+"/categories", tt.body).WithURLParam("id", tt.folderID)
			resp := helpers.MakeRequest(t, req, handler.UpdateFolderCategories)
			assert.Equal(t, tt.wantCode, resp.Code)

			if tt.wantCode == http.StatusOK {
				assert.NotNil(t, resp.Body["categories"])
			}
		})
	}
}
