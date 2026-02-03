package scans

import (
	"net/http"
	"testing"

	"stl-manager/tests/integration/helpers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetScanPaths(t *testing.T) {
	scan := helpers.CreateTestScan(t, "completed")
	defer helpers.DeleteTestScan(t, scan.ID)

	tests := []struct {
		name     string
		id       string
		wantCode int
	}{
		{
			name:     "get paths for existing scan",
			id:       scan.ID,
			wantCode: http.StatusOK,
		},
		{
			name:     "get paths for non-existent scan returns empty array",
			id:       uuid.New().String(),
			wantCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := helpers.GET("/scans/" + tt.id + "/paths").WithURLParam("id", tt.id)
			resp := helpers.MakeRequestArray(t, req, handler.GetScanPaths)
			assert.Equal(t, tt.wantCode, resp.Code)

			// Response should always be an array
			assert.NotNil(t, resp.Body)
		})
	}
}

func TestGetScanPathsResponseStructure(t *testing.T) {
	scan := helpers.CreateTestScan(t, "completed")
	defer helpers.DeleteTestScan(t, scan.ID)

	req := helpers.GET("/scans/" + scan.ID + "/paths").WithURLParam("id", scan.ID)
	resp := helpers.MakeRequestArray(t, req, handler.GetScanPaths)

	assert.Equal(t, http.StatusOK, resp.Code)

	// If there are paths, verify structure
	if len(resp.Body) > 0 {
		path := resp.Body[0]
		assert.NotNil(t, path["id"])
		assert.NotNil(t, path["scan_id"])
		assert.NotNil(t, path["root_path"])
		assert.NotNil(t, path["files_found"])
		assert.NotNil(t, path["files_inserted"])
		assert.NotNil(t, path["folders_found"])
		assert.NotNil(t, path["folders_inserted"])
		assert.NotNil(t, path["created_at"])
	}
}
