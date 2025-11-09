package categories

import (
	"net/http"

	"stl-manager/internal/db"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

func (h *Handler) SoftDeleteCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := db.New(h.pool)

	// Get ID from URL parameter
	idStr := chi.URLParam(r, "id")
	categoryID, err := uuid.Parse(idStr)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	// Soft delete category
	err = queries.SoftDeleteCategory(ctx, pgtype.UUID{Bytes: categoryID, Valid: true})
	if err != nil {
		h.logger.Error("failed to delete category", zap.Error(err))
		h.RespondError(w, http.StatusInternalServerError, "failed to delete category")
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]string{"message": "category deleted successfully"})
}

func (h *Handler) RestoreCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := db.New(h.pool)

	// Get ID from URL parameter
	idStr := chi.URLParam(r, "id")
	categoryID, err := uuid.Parse(idStr)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	// Restore category
	err = queries.RestoreCategory(ctx, pgtype.UUID{Bytes: categoryID, Valid: true})
	if err != nil {
		h.logger.Error("failed to restore category", zap.Error(err))
		h.RespondError(w, http.StatusInternalServerError, "failed to restore category")
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]string{"message": "category restored successfully"})
}
