package folders

import (
	"net/http"
	"testing"

	"stl-manager/tests/integration/helpers"

	"github.com/stretchr/testify/assert"
)

func TestListFolders(t *testing.T) {
	tests := []struct {
		name string
		req  helpers.HTTPTestRequest
		want int
	}{
		{
			name: "list all folders",
			req:  helpers.GET("/folders"),
			want: http.StatusOK,
		},
		{
			name: "list with pagination",
			req:  helpers.GET("/folders").WithQueryParam("page", "1").WithQueryParam("page_size", "10"),
			want: http.StatusOK,
		},
		{
			name: "search folders",
			req:  helpers.GET("/folders").WithQueryParam("q", "test"),
			want: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := helpers.MakeRequest(t, tt.req, handler.ListFolders)
			assert.Equal(t, tt.want, resp.Code)
			helpers.AssertPaginatedResponse(t, resp)
		})
	}
}
