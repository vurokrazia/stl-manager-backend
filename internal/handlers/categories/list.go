package categories

import (
	"encoding/json"
	"net/http"
	"strconv"

	"stl-manager/internal/db"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Handler struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func New(pool *pgxpool.Pool, logger *zap.Logger) *Handler {
	return &Handler{pool: pool, logger: logger}
}

func (h *Handler) RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (h *Handler) RespondError(w http.ResponseWriter, status int, message string) {
	h.RespondJSON(w, status, map[string]string{"error": message})
}

func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := db.New(h.pool)

	query := r.URL.Query()
	searchQuery := query.Get("q")
	page := 1
	pageSize := 20
	if p := query.Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := query.Get("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}
	offset := (page - 1) * pageSize

	var categories []db.Category
	var total int64
	var err error

	// Use search if query parameter provided
	if searchQuery != "" {
		categories, err = queries.SearchCategoriesPaginated(ctx, db.SearchCategoriesPaginatedParams{
			Search: searchQuery,
			Limit:  int32(pageSize),
			Offset: int32(offset),
		})
		if err != nil {
			h.logger.Error("failed to search categories", zap.Error(err))
			h.RespondError(w, http.StatusInternalServerError, "failed to search categories")
			return
		}

		total, err = queries.CountSearchCategories(ctx, searchQuery)
		if err != nil {
			h.logger.Error("failed to count search categories", zap.Error(err))
			total = 0
		}
	} else {
		categories, err = queries.ListCategoriesPaginated(ctx, db.ListCategoriesPaginatedParams{
			Limit:  int32(pageSize),
			Offset: int32(offset),
		})
		if err != nil {
			h.logger.Error("failed to list categories", zap.Error(err))
			h.RespondError(w, http.StatusInternalServerError, "failed to list categories")
			return
		}

		total, err = queries.CountCategories(ctx)
		if err != nil {
			h.logger.Error("failed to count categories", zap.Error(err))
			total = 0
		}
	}

	h.RespondJSON(w, http.StatusOK, map[string]any{
		"items":       categories,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": int((total + int64(pageSize) - 1) / int64(pageSize)),
	})
}
