package health

import (
	"net/http"
	"testing"

	"stl-manager/tests/integration/helpers"

	"github.com/stretchr/testify/assert"
)

func TestHealth(t *testing.T) {
	req := helpers.GET("/health")
	resp := helpers.MakeRequest(t, req, handler.Health)

	assert.Equal(t, http.StatusOK, resp.Code)
	helpers.AssertHasFields(t, resp.Body, "status", "service")
}

func TestGetAIStatus(t *testing.T) {
	req := helpers.GET("/ai/status")
	resp := helpers.MakeRequest(t, req, handler.GetAIStatus)

	assert.Equal(t, http.StatusOK, resp.Code)
	helpers.AssertHasFields(t, resp.Body, "enabled")
}
