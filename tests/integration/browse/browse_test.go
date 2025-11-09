package browse

import (
	"net/http"
	"testing"

	"stl-manager/tests/integration/helpers"

	"github.com/stretchr/testify/assert"
)

func TestListBrowse(t *testing.T) {
	tests := []struct {
		name string
		req  helpers.HTTPTestRequest
		want int
	}{
		{
			name: "list root folders",
			req:  helpers.GET("/browse"),
			want: http.StatusOK,
		},
		{
			name: "search folders",
			req:  helpers.GET("/browse").WithQueryParam("q", "test"),
			want: http.StatusOK,
		},
		{
			name: "with pagination",
			req:  helpers.GET("/browse").WithQueryParam("page", "1").WithQueryParam("page_size", "10"),
			want: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := helpers.MakeRequest(t, tt.req, handler.ListBrowse)
			assert.Equal(t, tt.want, resp.Code)
			helpers.AssertPaginatedResponse(t, resp)
		})
	}
}

func TestListMixed(t *testing.T) {
	tests := []struct {
		name string
		req  helpers.HTTPTestRequest
		want int
	}{
		{
			name: "list root mixed",
			req:  helpers.GET("/mixed"),
			want: http.StatusOK,
		},
		{
			name: "with pagination",
			req:  helpers.GET("/mixed").WithQueryParam("page", "1").WithQueryParam("page_size", "10"),
			want: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := helpers.MakeRequest(t, tt.req, handler.ListMixed)
			assert.Equal(t, tt.want, resp.Code)
			helpers.AssertPaginatedResponse(t, resp)
		})
	}
}
