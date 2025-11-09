package folders

import (
	"net/http"
	"testing"

	"stl-manager/tests/integration/helpers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetFolder(t *testing.T) {
	folder := helpers.CreateTestFolder(t, "test-folder")
	defer helpers.DeleteTestFolder(t, folder.ID)

	// Create some test files in the folder
	file1 := helpers.CreateTestFile(t, "file1", "stl", folder.ID)
	defer helpers.DeleteTestFile(t, file1.ID)

	file2 := helpers.CreateTestFile(t, "file2", "zip", folder.ID)
	defer helpers.DeleteTestFile(t, file2.ID)

	tests := []struct {
		name     string
		id       string
		wantCode int
	}{
		{
			name:     "get existing folder",
			id:       uuid.UUID(folder.ID.Bytes).String(),
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
			req := helpers.GET("/folders/" + tt.id).WithURLParam("id", tt.id)
			resp := helpers.MakeRequest(t, req, handler.GetFolder)
			assert.Equal(t, tt.wantCode, resp.Code)

			if tt.wantCode == http.StatusOK {
				assert.NotNil(t, resp.Body["folder"])
				assert.NotNil(t, resp.Body["files"])
				assert.NotNil(t, resp.Body["subfolders"])
				assert.NotNil(t, resp.Body["categories"])
				assert.NotNil(t, resp.Body["pagination"])
			}
		})
	}
}

func TestGetFolderWithPagination(t *testing.T) {
	folder := helpers.CreateTestFolder(t, "test-pagination")
	defer helpers.DeleteTestFolder(t, folder.ID)

	// Create multiple test files
	for i := 0; i < 5; i++ {
		file := helpers.CreateTestFile(t, "file-"+string(rune(i+'0')), "stl", folder.ID)
		defer helpers.DeleteTestFile(t, file.ID)
	}

	tests := []struct {
		name     string
		page     string
		pageSize string
		wantCode int
	}{
		{
			name:     "default pagination",
			wantCode: http.StatusOK,
		},
		{
			name:     "custom page size",
			page:     "1",
			pageSize: "2",
			wantCode: http.StatusOK,
		},
		{
			name:     "page 2",
			page:     "2",
			pageSize: "2",
			wantCode: http.StatusOK,
		},
	}

	folderID := uuid.UUID(folder.ID.Bytes).String()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := helpers.GET("/folders/" + folderID).WithURLParam("id", folderID)
			if tt.page != "" {
				req = req.WithQueryParam("page", tt.page)
			}
			if tt.pageSize != "" {
				req = req.WithQueryParam("page_size", tt.pageSize)
			}

			resp := helpers.MakeRequest(t, req, handler.GetFolder)
			assert.Equal(t, tt.wantCode, resp.Code)

			if tt.wantCode == http.StatusOK {
				pagination := resp.GetMap("pagination")
				assert.NotNil(t, pagination)
				assert.NotNil(t, pagination["total"])
				assert.NotNil(t, pagination["page"])
				assert.NotNil(t, pagination["page_size"])
			}
		})
	}
}
