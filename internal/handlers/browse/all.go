package browse

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"stl-manager/internal/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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

// ListBrowse returns a mixed list of folders and root-level files
func (h *Handler) ListBrowse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := db.New(h.pool)

	query := r.URL.Query()
	searchQuery := strings.TrimSpace(query.Get("q"))
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

	var folders []db.Folder
	var totalFolders int64
	var err error

	// Use search queries if search parameter is provided
	if searchQuery != "" {
		folders, err = queries.SearchRootFoldersPaginated(ctx, db.SearchRootFoldersPaginatedParams{
			Search: searchQuery,
			Limit:  int32(pageSize),
			Offset: int32(offset),
		})
		if err != nil {
			h.logger.Error("failed to search root folders", zap.Error(err))
			h.RespondError(w, http.StatusInternalServerError, "Failed to search folders")
			return
		}

		totalFolders, err = queries.CountSearchRootFolders(ctx, searchQuery)
		if err != nil {
			h.logger.Error("failed to count search root folders", zap.Error(err))
			totalFolders = 0
		}
	} else {
		folders, err = queries.ListRootFoldersPaginated(ctx, db.ListRootFoldersPaginatedParams{
			Limit:  int32(pageSize),
			Offset: int32(offset),
		})
		if err != nil {
			h.logger.Error("failed to list root folders", zap.Error(err))
			h.RespondError(w, http.StatusInternalServerError, "Failed to list folders")
			return
		}

		totalFolders, err = queries.CountRootFolders(ctx)
		if err != nil {
			h.logger.Error("failed to count folders", zap.Error(err))
			totalFolders = 0
		}
	}

	type BrowseItem struct {
		ID         string        `json:"id"`
		Name       string        `json:"name"`
		Type       string        `json:"type"`
		FileCount  *int          `json:"file_count,omitempty"`
		Categories []db.Category `json:"categories"`
		CreatedAt  string        `json:"created_at"`
	}

	items := make([]BrowseItem, 0, len(folders))

	for _, folder := range folders {
		// FIX: Use CountFolderFiles instead of loading all files
		fileCount, err := queries.CountFolderFiles(ctx, folder.ID)
		if err != nil {
			h.logger.Warn("failed to count folder files", zap.Error(err))
			fileCount = 0
		}
		count := int(fileCount)

		categories, err := queries.GetFolderCategories(ctx, folder.ID)
		if err != nil {
			h.logger.Warn("failed to get folder categories", zap.Error(err))
			categories = []db.Category{}
		}

		items = append(items, BrowseItem{
			ID:         uuid.UUID(folder.ID.Bytes).String(),
			Name:       folder.Name,
			Type:       "folder",
			FileCount:  &count,
			Categories: categories,
			CreatedAt:  folder.CreatedAt.Time.Format(time.RFC3339),
		})
	}

	h.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"items":       items,
		"total":       totalFolders,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": int((totalFolders + int64(pageSize) - 1) / int64(pageSize)),
	})
}

// ListMixed returns folders + files respecting hierarchy
func (h *Handler) ListMixed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := db.New(h.pool)

	query := r.URL.Query()
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

	folderIDStr := query.Get("folder_id")

	var folders []db.Folder
	var files []db.File
	var totalFolders int64
	var totalFiles int64
	var err error

	if folderIDStr == "" {
		folders, err = queries.ListRootFoldersPaginated(ctx, db.ListRootFoldersPaginatedParams{
			Limit:  int32(pageSize),
			Offset: int32(offset),
		})
		if err != nil {
			h.logger.Error("failed to list root folders", zap.Error(err))
			h.RespondError(w, http.StatusInternalServerError, "Failed to list folders")
			return
		}

		files, err = queries.ListRootFilesPaginated(ctx, db.ListRootFilesPaginatedParams{
			Limit:  int32(pageSize),
			Offset: int32(offset),
		})
		if err != nil {
			h.logger.Error("failed to list root files", zap.Error(err))
			h.RespondError(w, http.StatusInternalServerError, "Failed to list files")
			return
		}

		totalFolders, _ = queries.CountRootFolders(ctx)
		totalFiles, _ = queries.CountRootFiles(ctx)
	} else {
		folderUUID, err := uuid.Parse(folderIDStr)
		if err != nil {
			h.RespondError(w, http.StatusBadRequest, "Invalid folder_id")
			return
		}

		folders, err = queries.ListSubfolders(ctx, pgtype.UUID{Bytes: folderUUID, Valid: true})
		if err != nil {
			h.logger.Error("failed to list subfolders", zap.Error(err))
			h.RespondError(w, http.StatusInternalServerError, "Failed to list subfolders")
			return
		}

		files, err = queries.GetFolderFiles(ctx, pgtype.UUID{Bytes: folderUUID, Valid: true})
		if err != nil {
			h.logger.Error("failed to list folder files", zap.Error(err))
			h.RespondError(w, http.StatusInternalServerError, "Failed to list files")
			return
		}

		totalFiles = int64(len(files))
		totalFolders = int64(len(folders))
	}

	type MixedItem struct {
		ID         string        `json:"id"`
		Name       string        `json:"name"`
		Type       string        `json:"type"`
		Size       *int64        `json:"size,omitempty"`
		FileCount  *int          `json:"file_count,omitempty"`
		Categories []db.Category `json:"categories"`
		CreatedAt  string        `json:"created_at"`
	}

	items := make([]MixedItem, 0, len(folders)+len(files))

	for _, folder := range folders {
		// FIX: Use CountFolderFiles instead of loading all files
		fileCount, err := queries.CountFolderFiles(ctx, folder.ID)
		if err != nil {
			h.logger.Warn("failed to count folder files", zap.Error(err))
			fileCount = 0
		}
		count := int(fileCount)

		categories, err := queries.GetFolderCategories(ctx, folder.ID)
		if err != nil {
			h.logger.Warn("failed to get folder categories", zap.Error(err))
			categories = []db.Category{}
		}

		items = append(items, MixedItem{
			ID:         uuid.UUID(folder.ID.Bytes).String(),
			Name:       folder.Name,
			Type:       "folder",
			FileCount:  &count,
			Categories: categories,
			CreatedAt:  folder.CreatedAt.Time.Format(time.RFC3339),
		})
	}

	for _, file := range files {
		categories, err := queries.GetFileCategories(ctx, file.ID)
		if err != nil {
			h.logger.Warn("failed to get file categories", zap.Error(err))
			categories = []db.Category{}
		}

		items = append(items, MixedItem{
			ID:         uuid.UUID(file.ID.Bytes).String(),
			Name:       file.FileName,
			Type:       file.Type,
			Size:       &file.Size,
			Categories: categories,
			CreatedAt:  file.CreatedAt.Time.Format(time.RFC3339),
		})
	}

	h.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"items":       items,
		"total":       totalFolders + totalFiles,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": int((totalFolders + totalFiles + int64(pageSize) - 1) / int64(pageSize)),
	})
}
