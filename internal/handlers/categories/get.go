package categories

import (
	"net/http"

	"stl-manager/internal/db"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

func (h *Handler) GetCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := db.New(h.pool)

	// Get ID from URL parameter
	idStr := chi.URLParam(r, "id")
	categoryID, err := uuid.Parse(idStr)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	// Get category
	category, err := queries.GetCategory(ctx, pgtype.UUID{Bytes: categoryID, Valid: true})
	if err != nil {
		h.logger.Error("failed to get category", zap.Error(err))
		h.RespondError(w, http.StatusNotFound, "category not found")
		return
	}

	h.RespondJSON(w, http.StatusOK, category)
}
