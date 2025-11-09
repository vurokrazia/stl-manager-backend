package categories

import (
	"encoding/json"
	"net/http"

	"stl-manager/internal/db"

	"go.uber.org/zap"
)

type CreateCategoryRequest struct {
	Name string `json:"name"`
}

func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := db.New(h.pool)

	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate
	if req.Name == "" {
		h.RespondError(w, http.StatusBadRequest, "name is required")
		return
	}

	// Create category
	category, err := queries.CreateCategory(ctx, req.Name)
	if err != nil {
		h.logger.Error("failed to create category", zap.Error(err))
		h.RespondError(w, http.StatusInternalServerError, "failed to create category")
		return
	}

	h.RespondJSON(w, http.StatusCreated, category)
}
