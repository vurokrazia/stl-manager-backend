package scans

import (
	"net/http"
	"testing"

	"stl-manager/tests/integration/helpers"

	"github.com/stretchr/testify/assert"
)

func TestListScans(t *testing.T) {
	tests := []struct {
		name string
		req  helpers.HTTPTestRequest
		want int
	}{
		{
			name: "list all scans",
			req:  helpers.GET("/scans"),
			want: http.StatusOK,
		},
		{
			name: "with pagination",
			req:  helpers.GET("/scans").WithQueryParam("page", "1").WithQueryParam("page_size", "10"),
			want: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := helpers.MakeRequest(t, tt.req, handler.ListScans)
			assert.Equal(t, tt.want, resp.Code)
			helpers.AssertPaginatedResponse(t, resp)
		})
	}
}
