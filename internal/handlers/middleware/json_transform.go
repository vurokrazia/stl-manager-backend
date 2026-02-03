package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"unicode"
)

// responseWriter wraps http.ResponseWriter to capture the response
type responseWriter struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.body.Write(b)
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
}

// JSONTransform middleware transforms PascalCase JSON keys to camelCase
func JSONTransform(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a response writer wrapper
		rw := &responseWriter{
			ResponseWriter: w,
			body:           &bytes.Buffer{},
			statusCode:     http.StatusOK,
		}

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Check if response is JSON
		contentType := w.Header().Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			// Not JSON, write original response
			w.WriteHeader(rw.statusCode)
			w.Write(rw.body.Bytes())
			return
		}

		// Parse and transform JSON
		var data interface{}
		if err := json.Unmarshal(rw.body.Bytes(), &data); err != nil {
			// Failed to parse, write original response
			w.WriteHeader(rw.statusCode)
			w.Write(rw.body.Bytes())
			return
		}

		// Transform keys
		transformed := transformKeys(data)

		// Re-encode
		result, err := json.Marshal(transformed)
		if err != nil {
			// Failed to encode, write original response
			w.WriteHeader(rw.statusCode)
			w.Write(rw.body.Bytes())
			return
		}

		// Write transformed response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(rw.statusCode)
		w.Write(result)
	})
}

// transformKeys recursively transforms all map keys from PascalCase to camelCase
func transformKeys(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			newKey := toCamelCase(key)
			result[newKey] = transformKeys(value)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = transformKeys(item)
		}
		return result
	default:
		return data
	}
}

// toCamelCase converts a string from PascalCase to camelCase
func toCamelCase(s string) string {
	if s == "" {
		return s
	}

	// Handle special cases
	switch s {
	case "ID":
		return "id"
	case "URL":
		return "url"
	case "API":
		return "api"
	}

	// Handle keys that end with ID (like FolderID, ParentFolderID)
	if strings.HasSuffix(s, "ID") && len(s) > 2 {
		base := s[:len(s)-2]
		return toCamelCase(base) + "Id"
	}

	// Standard conversion: lowercase first letter
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}
