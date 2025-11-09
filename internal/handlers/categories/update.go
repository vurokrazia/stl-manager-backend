package categories

import (
	"encoding/json"
	"net/http"

	"stl-manager/internal/db"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type UpdateCategoryRequest struct {
	Name string `json:"name"`
}

func (h *Handler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := db.New(h.pool)

	// Get ID from URL parameter
	idStr := chi.URLParam(r, "id")
	categoryID, err := uuid.Parse(idStr)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	// Parse request body
	var req UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate
	if req.Name == "" {
		h.RespondError(w, http.StatusBadRequest, "name is required")
		return
	}

	// Update category
	category, err := queries.UpdateCategory(ctx, db.UpdateCategoryParams{
		ID:   pgtype.UUID{Bytes: categoryID, Valid: true},
		Name: req.Name,
	})
	if err != nil {
		h.logger.Error("failed to update category", zap.Error(err))
		h.RespondError(w, http.StatusInternalServerError, "failed to update category")
		return
	}

	h.RespondJSON(w, http.StatusOK, category)
}
