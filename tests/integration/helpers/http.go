package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

// HTTPTestRequest represents a test HTTP request
type HTTPTestRequest struct {
	Method      string
	URL         string
	Body        interface{}
	URLParams   map[string]string
	QueryParams map[string]string
	Headers     map[string]string
}

// HTTPTestResponse represents a test HTTP response
type HTTPTestResponse struct {
	*httptest.ResponseRecorder
	Body map[string]interface{}
}

// MakeRequest creates and executes an HTTP test request
func MakeRequest(t *testing.T, req HTTPTestRequest, handler http.HandlerFunc) *HTTPTestResponse {
	// Prepare request body
	var bodyReader io.Reader
	if req.Body != nil {
		if str, ok := req.Body.(string); ok {
			bodyReader = bytes.NewBufferString(str)
		} else {
			bodyBytes, err := json.Marshal(req.Body)
			require.NoError(t, err, "Failed to marshal request body")
			bodyReader = bytes.NewBuffer(bodyBytes)
		}
	}

	// Create HTTP request
	httpReq := httptest.NewRequest(req.Method, req.URL, bodyReader)

	// Add headers
	if req.Headers != nil {
		for key, value := range req.Headers {
			httpReq.Header.Set(key, value)
		}
	}

	// Add default Content-Type if not set and body exists
	if req.Body != nil && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// Add URL params (for chi router)
	if req.URLParams != nil {
		rctx := chi.NewRouteContext()
		for key, value := range req.URLParams {
			rctx.URLParams.Add(key, value)
		}
		httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))
	}

	// Add query params
	if req.QueryParams != nil {
		q := httpReq.URL.Query()
		for key, value := range req.QueryParams {
			q.Add(key, value)
		}
		httpReq.URL.RawQuery = q.Encode()
	}

	// Execute request
	recorder := httptest.NewRecorder()
	handler(recorder, httpReq)

	// Parse response body
	response := &HTTPTestResponse{
		ResponseRecorder: recorder,
		Body:             make(map[string]interface{}),
	}

	if recorder.Body.Len() > 0 {
		err := json.NewDecoder(recorder.Body).Decode(&response.Body)
		if err != nil {
			t.Logf("Warning: failed to decode response body: %v", err)
		}
	}

	return response
}

// GET creates a GET request
func GET(url string) HTTPTestRequest {
	return HTTPTestRequest{
		Method: http.MethodGet,
		URL:    url,
	}
}

// POST creates a POST request
func POST(url string, body interface{}) HTTPTestRequest {
	return HTTPTestRequest{
		Method: http.MethodPost,
		URL:    url,
		Body:   body,
	}
}

// PUT creates a PUT request
func PUT(url string, body interface{}) HTTPTestRequest {
	return HTTPTestRequest{
		Method: http.MethodPut,
		URL:    url,
		Body:   body,
	}
}

// PATCH creates a PATCH request
func PATCH(url string, body interface{}) HTTPTestRequest {
	return HTTPTestRequest{
		Method: http.MethodPatch,
		URL:    url,
		Body:   body,
	}
}

// DELETE creates a DELETE request
func DELETE(url string) HTTPTestRequest {
	return HTTPTestRequest{
		Method: http.MethodDelete,
		URL:    url,
	}
}

// WithURLParam adds a URL parameter to the request
func (r HTTPTestRequest) WithURLParam(key, value string) HTTPTestRequest {
	if r.URLParams == nil {
		r.URLParams = make(map[string]string)
	}
	r.URLParams[key] = value
	return r
}

// WithQueryParam adds a query parameter to the request
func (r HTTPTestRequest) WithQueryParam(key, value string) HTTPTestRequest {
	if r.QueryParams == nil {
		r.QueryParams = make(map[string]string)
	}
	r.QueryParams[key] = value
	return r
}

// WithHeader adds a header to the request
func (r HTTPTestRequest) WithHeader(key, value string) HTTPTestRequest {
	if r.Headers == nil {
		r.Headers = make(map[string]string)
	}
	r.Headers[key] = value
	return r
}

// Response helper methods

// GetString gets a string value from response body
func (r *HTTPTestResponse) GetString(key string) string {
	if val, ok := r.Body[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// GetFloat gets a float64 value from response body
func (r *HTTPTestResponse) GetFloat(key string) float64 {
	if val, ok := r.Body[key]; ok {
		if num, ok := val.(float64); ok {
			return num
		}
	}
	return 0
}

// GetArray gets an array value from response body
func (r *HTTPTestResponse) GetArray(key string) []interface{} {
	if val, ok := r.Body[key]; ok {
		if arr, ok := val.([]interface{}); ok {
			return arr
		}
	}
	return nil
}

// GetMap gets a map value from response body
func (r *HTTPTestResponse) GetMap(key string) map[string]interface{} {
	if val, ok := r.Body[key]; ok {
		if m, ok := val.(map[string]interface{}); ok {
			return m
		}
	}
	return nil
}
