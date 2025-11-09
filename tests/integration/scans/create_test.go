package scans

import (
	"net/http"
	"testing"

	"stl-manager/tests/integration/helpers"

	"github.com/stretchr/testify/assert"
)

func TestCreateScan(t *testing.T) {
	tests := []struct {
		name     string
		wantCode int
	}{
		{
			name:     "create scan successfully",
			wantCode: http.StatusAccepted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := helpers.POST("/scan", nil)
			resp := helpers.MakeRequest(t, req, handler.CreateScan)
			assert.Equal(t, tt.wantCode, resp.Code)

			if tt.wantCode == http.StatusAccepted {
				scanID := resp.GetString("scan_id")
				assert.NotEmpty(t, scanID, "scan_id should be returned")

				// Note: We don't clean up the scan here because the goroutine is running
				// In a real scenario, scans would be cleaned up by a background job or TTL
			}
		})
	}
}
