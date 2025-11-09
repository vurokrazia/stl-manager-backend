package files

import (
	"net/http"
	"testing"

	"stl-manager/tests/integration/helpers"

	"github.com/stretchr/testify/assert"
)

func TestListFiles(t *testing.T) {
	tests := []struct {
		name string
		req  helpers.HTTPTestRequest
		want int
	}{
		{
			name: "list all files",
			req:  helpers.GET("/files"),
			want: http.StatusOK,
		},
		{
			name: "list with pagination",
			req:  helpers.GET("/files").WithQueryParam("page", "1").WithQueryParam("page_size", "10"),
			want: http.StatusOK,
		},
		{
			name: "search files",
			req:  helpers.GET("/files").WithQueryParam("q", "test"),
			want: http.StatusOK,
		},
		{
			name: "filter by type",
			req:  helpers.GET("/files").WithQueryParam("type", "stl"),
			want: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := helpers.MakeRequest(t, tt.req, handler.ListFiles)
			assert.Equal(t, tt.want, resp.Code)
			helpers.AssertPaginatedResponse(t, resp)
		})
	}
}
