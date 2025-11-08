package handlers

import (
	"net/http"
)

// Health check
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check database connection
	if err := h.pool.Ping(ctx); err != nil {
		h.respondError(w, http.StatusServiceUnavailable, "database unhealthy")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "stl-manager-api",
	})
}

// GetAIStatus returns whether AI classification is enabled
func (h *Handler) GetAIStatus(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, map[string]any{
		"enabled": h.classifier.IsEnabled(),
	})
}
