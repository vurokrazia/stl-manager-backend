package files

import (
	"net/http"
	"strconv"

	"stl-manager/internal/db"

	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

func (h *Handler) ListFiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := db.New(h.pool)

	// Parse query parameters
	query := r.URL.Query()
	searchQuery := query.Get("q")
	typeFilter := query.Get("type")
	categoryFilter := query.Get("category")

	// Parse pagination
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

	var files []db.File
	var total int64
	var err error

	// If search query or type filter, use SearchFiles
	if searchQuery != "" || typeFilter != "" {
		searchRows, err := queries.SearchFiles(ctx, db.SearchFilesParams{
			Similarity: searchQuery,
			Column2:    typeFilter,
			Limit:      int32(pageSize),
			Offset:     int32(offset),
		})
		if err != nil {
			h.logger.Error("failed to search files", zap.Error(err))
			h.RespondError(w, http.StatusInternalServerError, "failed to search files")
			return
		}

		// Convert SearchFilesRow to File
		files = make([]db.File, len(searchRows))
		for i, row := range searchRows {
			files[i] = db.File{
				ID:         row.ID,
				Path:       row.Path,
				FileName:   row.FileName,
				Type:       row.Type,
				Size:       row.Size,
				ModifiedAt: row.ModifiedAt,
				Sha256:     row.Sha256,
				CreatedAt:  row.CreatedAt,
				UpdatedAt:  row.UpdatedAt,
			}
		}

		// Count total (approximate for search)
		if typeFilter != "" {
			total, _ = queries.CountFilesByType(ctx, typeFilter)
		} else {
			total, _ = queries.CountFiles(ctx)
		}
	} else if categoryFilter != "" {
		// Filter by category
		files, err = queries.GetFilesByCategory(ctx, db.GetFilesByCategoryParams{
			Name:   categoryFilter,
			Limit:  int32(pageSize),
			Offset: int32(offset),
		})
		if err != nil {
			h.logger.Error("failed to get files by category", zap.Error(err))
			h.RespondError(w, http.StatusInternalServerError, "failed to get files by category")
			return
		}
		total, _ = queries.CountFiles(ctx)
	} else {
		// Default: list all files
		files, err = queries.ListFiles(ctx, db.ListFilesParams{
			Limit:  int32(pageSize),
			Offset: int32(offset),
		})
		if err != nil {
			h.logger.Error("failed to list files", zap.Error(err))
			h.RespondError(w, http.StatusInternalServerError, "failed to list files")
			return
		}
		total, _ = queries.CountFiles(ctx)
	}

	// Attach categories to each file using batch query (1 query instead of N)
	type FileWithCategories struct {
		db.File
		Categories []db.Category `json:"categories"`
	}

	// Collect file IDs
	fileIDs := make([]pgtype.UUID, len(files))
	for i, file := range files {
		fileIDs[i] = file.ID
	}

	// Get all categories in one query
	categoriesMap := make(map[pgtype.UUID][]db.Category)
	if len(fileIDs) > 0 {
		batchResults, err := queries.GetCategoriesBatch(ctx, fileIDs)
		if err != nil {
			h.logger.Warn("failed to get file categories batch", zap.Error(err))
		} else {
			// Group categories by file_id
			for _, row := range batchResults {
				categoriesMap[row.FileID] = append(categoriesMap[row.FileID], db.Category{
					ID:        row.ID,
					Name:      row.Name,
					CreatedAt: row.CreatedAt,
				})
			}
		}
	}

	// Build response with categories
	filesWithCategories := make([]FileWithCategories, len(files))
	for i, file := range files {
		categories := categoriesMap[file.ID]
		if categories == nil {
			categories = []db.Category{}
		}
		filesWithCategories[i] = FileWithCategories{
			File:       file,
			Categories: categories,
		}
	}

	h.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"items":       filesWithCategories,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": int((total + int64(pageSize) - 1) / int64(pageSize)),
	})
}
