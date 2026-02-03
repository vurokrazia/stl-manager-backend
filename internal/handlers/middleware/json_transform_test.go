package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONTransform(t *testing.T) {
	tests := []struct {
		name           string
		inputJSON      string
		expectedJSON   string
		contentType    string
		skipTransform  bool
	}{
		{
			name:         "transforms PascalCase to camelCase",
			inputJSON:    `{"UserName":"test","Email":"test@example.com"}`,
			expectedJSON: `{"email":"test@example.com","userName":"test"}`,
			contentType:  "application/json",
		},
		{
			name:         "transforms ID to id",
			inputJSON:    `{"ID":"123","Name":"Test"}`,
			expectedJSON: `{"id":"123","name":"Test"}`,
			contentType:  "application/json",
		},
		{
			name:         "transforms FolderID to folderId",
			inputJSON:    `{"FolderID":"abc","ParentFolderID":"def"}`,
			expectedJSON: `{"folderId":"abc","parentFolderId":"def"}`,
			contentType:  "application/json",
		},
		{
			name:         "handles nested objects",
			inputJSON:    `{"User":{"FirstName":"John","LastName":"Doe"}}`,
			expectedJSON: `{"user":{"firstName":"John","lastName":"Doe"}}`,
			contentType:  "application/json",
		},
		{
			name:         "handles arrays",
			inputJSON:    `[{"ID":"1","Name":"First"},{"ID":"2","Name":"Second"}]`,
			expectedJSON: `[{"id":"1","name":"First"},{"id":"2","name":"Second"}]`,
			contentType:  "application/json",
		},
		{
			name:          "skips non-JSON content",
			inputJSON:     `<html>Not JSON</html>`,
			expectedJSON:  `<html>Not JSON</html>`,
			contentType:   "text/html",
			skipTransform: true,
		},
		{
			name:         "handles empty object",
			inputJSON:    `{}`,
			expectedJSON: `{}`,
			contentType:  "application/json",
		},
		{
			name:         "handles empty array",
			inputJSON:    `[]`,
			expectedJSON: `[]`,
			contentType:  "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler that returns the test JSON
			innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tt.contentType)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.inputJSON))
			})

			// Wrap with JSONTransform middleware
			handler := JSONTransform(innerHandler)

			// Make request
			req := httptest.NewRequest("GET", "/", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, http.StatusOK, rec.Code)

			if tt.skipTransform {
				assert.Equal(t, tt.expectedJSON, rec.Body.String())
			} else {
				assert.JSONEq(t, tt.expectedJSON, rec.Body.String())
			}
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"ID", "id"},
		{"URL", "url"},
		{"API", "api"},
		{"Name", "name"},
		{"UserName", "userName"},
		{"FirstName", "firstName"},
		{"FolderID", "folderId"},
		{"ParentFolderID", "parentFolderId"},
		{"FileID", "fileId"},
		{"ScanID", "scanId"},
		{"CreatedAt", "createdAt"},
		{"UpdatedAt", "updatedAt"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toCamelCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTransformKeys(t *testing.T) {
	t.Run("transforms map keys", func(t *testing.T) {
		input := map[string]interface{}{
			"UserName": "test",
			"ID":       123,
		}

		result := transformKeys(input).(map[string]interface{})

		assert.Equal(t, "test", result["userName"])
		assert.Equal(t, 123, result["id"])
		assert.Nil(t, result["UserName"])
		assert.Nil(t, result["ID"])
	})

	t.Run("transforms nested maps", func(t *testing.T) {
		input := map[string]interface{}{
			"User": map[string]interface{}{
				"FirstName": "John",
			},
		}

		result := transformKeys(input).(map[string]interface{})
		user := result["user"].(map[string]interface{})

		assert.Equal(t, "John", user["firstName"])
	})

	t.Run("transforms arrays of maps", func(t *testing.T) {
		input := []interface{}{
			map[string]interface{}{"ID": 1},
			map[string]interface{}{"ID": 2},
		}

		result := transformKeys(input).([]interface{})

		assert.Len(t, result, 2)
		assert.Equal(t, 1, result[0].(map[string]interface{})["id"])
		assert.Equal(t, 2, result[1].(map[string]interface{})["id"])
	})

	t.Run("preserves primitive values", func(t *testing.T) {
		assert.Equal(t, "string", transformKeys("string"))
		assert.Equal(t, 123, transformKeys(123))
		assert.Equal(t, true, transformKeys(true))
		assert.Nil(t, transformKeys(nil))
	})
}
