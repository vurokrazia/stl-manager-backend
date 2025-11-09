package folders

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"stl-manager/internal/db"

	"github.com/go-chi/chi/v5"
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

// ListFolders lists all folders with their file count
func (h *Handler) ListFolders(w http.ResponseWriter, r *http.Request) {
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
	var total int64
	var err error

	// Use search queries if search parameter is provided
	if searchQuery != "" {
		folders, err = queries.SearchFoldersPaginated(ctx, db.SearchFoldersPaginatedParams{
			Search: searchQuery,
			Limit:  int32(pageSize),
			Offset: int32(offset),
		})
		if err != nil {
			h.logger.Error("failed to search folders", zap.Error(err))
			h.RespondError(w, http.StatusInternalServerError, "Failed to search folders")
			return
		}

		total, err = queries.CountSearchFolders(ctx, searchQuery)
		if err != nil {
			h.logger.Error("failed to count search folders", zap.Error(err))
			total = 0
		}
	} else {
		folders, err = queries.ListFoldersPaginated(ctx, db.ListFoldersPaginatedParams{
			Limit:  int32(pageSize),
			Offset: int32(offset),
		})
		if err != nil {
			h.logger.Error("failed to list folders", zap.Error(err))
			h.RespondError(w, http.StatusInternalServerError, "Failed to list folders")
			return
		}

		total, err = queries.CountFolders(ctx)
		if err != nil {
			h.logger.Error("failed to count folders", zap.Error(err))
			total = 0
		}
	}

	type FolderResponse struct {
		db.Folder
		FileCount  int           `json:"file_count"`
		Categories []db.Category `json:"categories"`
	}

	// Collect folder IDs for batch queries
	folderIDs := make([]pgtype.UUID, len(folders))
	for i, folder := range folders {
		folderIDs[i] = folder.ID
	}

	// Get all folder categories in one batch query (1 query instead of N)
	categoriesMap := make(map[pgtype.UUID][]db.Category)
	if len(folderIDs) > 0 {
		batchResults, err := queries.GetFolderCategoriesBatch(ctx, folderIDs)
		if err != nil {
			h.logger.Warn("failed to get folder categories batch", zap.Error(err))
		} else {
			for _, row := range batchResults {
				categoriesMap[row.FolderID] = append(categoriesMap[row.FolderID], db.Category{
					ID:        row.ID,
					Name:      row.Name,
					CreatedAt: row.CreatedAt,
				})
			}
		}
	}

	response := make([]FolderResponse, len(folders))
	for i, folder := range folders {
		// FIX: Use CountFolderFiles instead of loading all files
		fileCount, err := queries.CountFolderFiles(ctx, folder.ID)
		if err != nil {
			h.logger.Warn("failed to count folder files", zap.Error(err))
			fileCount = 0
		}

		categories := categoriesMap[folder.ID]
		if categories == nil {
			categories = []db.Category{}
		}

		response[i] = FolderResponse{
			Folder:     folder,
			FileCount:  int(fileCount),
			Categories: categories,
		}
	}

	h.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"items":       response,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": int((total + int64(pageSize) - 1) / int64(pageSize)),
	})
}

// GetFolder gets a specific folder with its files (paginated)
func (h *Handler) GetFolder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := db.New(h.pool)

	folderIDStr := chi.URLParam(r, "id")
	folderID, err := uuid.Parse(folderIDStr)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "Invalid folder ID")
		return
	}

	searchQuery := strings.TrimSpace(r.URL.Query().Get("search"))
	typeFilter := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("type")))
	categoryFilter := strings.TrimSpace(r.URL.Query().Get("category"))

	query := r.URL.Query()
	page := 1
	pageSize := 50
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

	folder, err := queries.GetFolder(ctx, pgtype.UUID{Bytes: folderID, Valid: true})
	if err != nil {
		h.logger.Error("failed to get folder", zap.Error(err))
		h.RespondError(w, http.StatusNotFound, "Folder not found")
		return
	}

	subfolders, err := queries.ListSubfolders(ctx, folder.ID)
	if err != nil {
		h.logger.Error("failed to get subfolders", zap.Error(err))
		h.RespondError(w, http.StatusInternalServerError, "Failed to get subfolders")
		return
	}

	if searchQuery != "" || categoryFilter != "" {
		filtered := []db.Folder{}
		searchLower := strings.ToLower(searchQuery)
		for _, subfolder := range subfolders {
			if searchQuery != "" && !strings.Contains(strings.ToLower(subfolder.Name), searchLower) {
				continue
			}
			if categoryFilter != "" {
				subfolderCategories, err := queries.GetFolderCategories(ctx, subfolder.ID)
				if err != nil || len(subfolderCategories) == 0 {
					continue
				}
				hasCategory := false
				for _, cat := range subfolderCategories {
					if cat.Name == categoryFilter {
						hasCategory = true
						break
					}
				}
				if !hasCategory {
					continue
				}
			}
			filtered = append(filtered, subfolder)
		}
		subfolders = filtered
	}

	var files []db.File
	var totalFiles int64
	hasFilters := searchQuery != "" || typeFilter != "" || categoryFilter != ""

	if hasFilters {
		allFiles, err := queries.GetFolderFiles(ctx, folder.ID)
		if err != nil {
			h.logger.Error("failed to get folder files", zap.Error(err))
			h.RespondError(w, http.StatusInternalServerError, "Failed to get folder files")
			return
		}

		filtered := []db.File{}
		searchLower := strings.ToLower(searchQuery)

		for _, file := range allFiles {
			if searchQuery != "" && !strings.Contains(strings.ToLower(file.FileName), searchLower) {
				continue
			}
			if typeFilter != "" && strings.ToLower(file.Type) != typeFilter {
				continue
			}
			if categoryFilter != "" {
				fileCategories, err := queries.GetFileCategories(ctx, file.ID)
				if err != nil || len(fileCategories) == 0 {
					continue
				}
				hasCategory := false
				for _, cat := range fileCategories {
					if cat.Name == categoryFilter {
						hasCategory = true
						break
					}
				}
				if !hasCategory {
					continue
				}
			}
			filtered = append(filtered, file)
		}

		totalFiles = int64(len(filtered))
		start := offset
		end := offset + pageSize
		if start > len(filtered) {
			start = len(filtered)
		}
		if end > len(filtered) {
			end = len(filtered)
		}
		files = filtered[start:end]
	} else {
		var err error
		files, err = queries.GetFolderFilesPaginated(ctx, db.GetFolderFilesPaginatedParams{
			FolderID: folder.ID,
			Limit:    int32(pageSize),
			Offset:   int32(offset),
		})
		if err != nil {
			h.logger.Error("failed to get folder files", zap.Error(err))
			h.RespondError(w, http.StatusInternalServerError, "Failed to get folder files")
			return
		}

		totalFiles, err = queries.CountFolderFiles(ctx, folder.ID)
		if err != nil {
			h.logger.Error("failed to count folder files", zap.Error(err))
			totalFiles = 0
		}
	}

	categories, err := queries.GetFolderCategories(ctx, folder.ID)
	if err != nil {
		h.logger.Warn("failed to get folder categories", zap.Error(err))
		categories = []db.Category{}
	}

	type SubfolderWithInfo struct {
		db.Folder
		FileCount  int           `json:"file_count"`
		Categories []db.Category `json:"categories"`
	}

	// Batch query for subfolder categories (1 query instead of N)
	subfolderCategoriesMap := make(map[pgtype.UUID][]db.Category)
	if len(subfolders) > 0 {
		subfolderIDs := make([]pgtype.UUID, len(subfolders))
		for i, subfolder := range subfolders {
			subfolderIDs[i] = subfolder.ID
		}
		batchResults, err := queries.GetFolderCategoriesBatch(ctx, subfolderIDs)
		if err != nil {
			h.logger.Warn("failed to get subfolder categories batch", zap.Error(err))
		} else {
			for _, row := range batchResults {
				subfolderCategoriesMap[row.FolderID] = append(subfolderCategoriesMap[row.FolderID], db.Category{
					ID:        row.ID,
					Name:      row.Name,
					CreatedAt: row.CreatedAt,
				})
			}
		}
	}

	subfoldersWithInfo := make([]SubfolderWithInfo, len(subfolders))
	for i, subfolder := range subfolders {
		// FIX: Use CountFolderFiles instead of loading all files
		fileCount, err := queries.CountFolderFiles(ctx, subfolder.ID)
		if err != nil {
			h.logger.Warn("failed to count subfolder files", zap.Error(err))
			fileCount = 0
		}

		subfolderCategories := subfolderCategoriesMap[subfolder.ID]
		if subfolderCategories == nil {
			subfolderCategories = []db.Category{}
		}

		subfoldersWithInfo[i] = SubfolderWithInfo{
			Folder:     subfolder,
			FileCount:  int(fileCount),
			Categories: subfolderCategories,
		}
	}

	type FileWithCategories struct {
		db.File
		Categories []db.Category `json:"categories"`
	}

	// Batch query for file categories (1 query instead of M)
	fileCategoriesMap := make(map[pgtype.UUID][]db.Category)
	if len(files) > 0 {
		fileIDs := make([]pgtype.UUID, len(files))
		for i, file := range files {
			fileIDs[i] = file.ID
		}
		batchResults, err := queries.GetCategoriesBatch(ctx, fileIDs)
		if err != nil {
			h.logger.Warn("failed to get file categories batch", zap.Error(err))
		} else {
			for _, row := range batchResults {
				fileCategoriesMap[row.FileID] = append(fileCategoriesMap[row.FileID], db.Category{
					ID:        row.ID,
					Name:      row.Name,
					CreatedAt: row.CreatedAt,
				})
			}
		}
	}

	filesWithCategories := make([]FileWithCategories, len(files))
	for i, file := range files {
		fileCategories := fileCategoriesMap[file.ID]
		if fileCategories == nil {
			fileCategories = []db.Category{}
		}
		filesWithCategories[i] = FileWithCategories{
			File:       file,
			Categories: fileCategories,
		}
	}

	h.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"folder":     folder,
		"subfolders": subfoldersWithInfo,
		"files":      filesWithCategories,
		"categories": categories,
		"pagination": map[string]interface{}{
			"total":       totalFiles,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": int((totalFiles + int64(pageSize) - 1) / int64(pageSize)),
		},
	})
}

// UpdateFolderCategories updates the categories assigned to a folder
func (h *Handler) UpdateFolderCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := db.New(h.pool)

	folderIDStr := chi.URLParam(r, "id")
	folderID, err := uuid.Parse(folderIDStr)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "Invalid folder ID")
		return
	}

	var req struct {
		CategoryIDs       []string `json:"category_ids"`
		ApplyToSTL        bool     `json:"apply_to_stl"`
		ApplyToZIP        bool     `json:"apply_to_zip"`
		ApplyToRAR        bool     `json:"apply_to_rar"`
		ApplyToSubfolders bool     `json:"apply_to_subfolders"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := queries.SetFolderCategories(ctx, pgtype.UUID{Bytes: folderID, Valid: true}); err != nil {
		h.logger.Error("failed to clear folder categories", zap.Error(err))
		h.RespondError(w, http.StatusInternalServerError, "Failed to update categories")
		return
	}

	categoryUUIDs := []pgtype.UUID{}
	for _, catIDStr := range req.CategoryIDs {
		catID, err := uuid.Parse(catIDStr)
		if err != nil {
			h.logger.Warn("invalid category ID", zap.String("id", catIDStr))
			continue
		}

		categoryUUID := pgtype.UUID{Bytes: catID, Valid: true}
		categoryUUIDs = append(categoryUUIDs, categoryUUID)

		err = queries.AddFolderCategory(ctx, db.AddFolderCategoryParams{
			FolderID:   pgtype.UUID{Bytes: folderID, Valid: true},
			CategoryID: categoryUUID,
		})
		if err != nil {
			h.logger.Error("failed to add folder category", zap.Error(err))
			continue
		}
	}

	if req.ApplyToSTL || req.ApplyToZIP || req.ApplyToRAR || req.ApplyToSubfolders {
		h.propagateFolderCategories(ctx, queries, pgtype.UUID{Bytes: folderID, Valid: true}, categoryUUIDs, req)
	}

	categories, err := queries.GetFolderCategories(ctx, pgtype.UUID{Bytes: folderID, Valid: true})
	if err != nil {
		h.logger.Error("failed to get updated categories", zap.Error(err))
		categories = []db.Category{}
	}

	h.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"categories": categories,
	})
}

func (h *Handler) propagateFolderCategories(
	ctx context.Context,
	queries *db.Queries,
	folderID pgtype.UUID,
	categoryUUIDs []pgtype.UUID,
	req struct {
		CategoryIDs       []string `json:"category_ids"`
		ApplyToSTL        bool     `json:"apply_to_stl"`
		ApplyToZIP        bool     `json:"apply_to_zip"`
		ApplyToRAR        bool     `json:"apply_to_rar"`
		ApplyToSubfolders bool     `json:"apply_to_subfolders"`
	},
) {
	// OPTIMIZED: Use batch operations instead of individual queries
	files, err := queries.GetFolderFiles(ctx, folderID)
	if err != nil {
		h.logger.Error("failed to get folder files for propagation", zap.Error(err))
		return
	}

	// Filter files that need category updates
	var filesToUpdate []pgtype.UUID
	for _, file := range files {
		shouldApply := false
		switch strings.ToLower(file.Type) {
		case "stl":
			shouldApply = req.ApplyToSTL
		case "zip":
			shouldApply = req.ApplyToZIP
		case "rar":
			shouldApply = req.ApplyToRAR
		}

		if shouldApply {
			filesToUpdate = append(filesToUpdate, file.ID)
		}
	}

	// Batch delete all categories for affected files (1 query instead of N)
	if len(filesToUpdate) > 0 {
		if err := queries.BulkRemoveFileCategories(ctx, filesToUpdate); err != nil {
			h.logger.Error("failed to bulk remove file categories", zap.Error(err))
			return
		}

		// Prepare bulk insert data
		var fileIDs []pgtype.UUID
		var catIDs []pgtype.UUID
		for _, fileID := range filesToUpdate {
			for _, catUUID := range categoryUUIDs {
				fileIDs = append(fileIDs, fileID)
				catIDs = append(catIDs, catUUID)
			}
		}

		// Batch insert all categories (1 query instead of N×M)
		if len(fileIDs) > 0 {
			if err := queries.BulkAddFileCategories(ctx, db.BulkAddFileCategoriesParams{
				FileIds:     fileIDs,
				CategoryIds: catIDs,
			}); err != nil {
				h.logger.Error("failed to bulk add file categories", zap.Error(err))
			}
		}

		h.logger.Info("bulk updated file categories",
			zap.Int("files", len(filesToUpdate)),
			zap.Int("categories", len(categoryUUIDs)))
	}

	// TODO: FEATURE FUTURA - Procesamiento recursivo en background
	// Implementar un job/worker que procese folders recursivamente sin bloquear el HTTP request
	// Esto permitirá aplicar categorías a toda la jerarquía de folders sin causar timeout
	// Consideraciones:
	// - Background job queue (Redis/RabbitMQ/Go channels)
	// - Progress tracking para mostrar al usuario
	// - Límite de profundidad configurable
	// - Transacciones por lotes para eficiencia
	// - Capacidad de cancelar la operación
	if req.ApplyToSubfolders {
		subfolders, err := queries.ListSubfolders(ctx, folderID)
		if err != nil {
			h.logger.Error("failed to get subfolders for propagation", zap.Error(err))
			return
		}

		// OPTIMIZED: Solo procesa subfolders del nivel actual (sin recursión)
		// Usar batch operations para actualizar múltiples folders a la vez
		if len(subfolders) > 0 {
			var subfolderIDs []pgtype.UUID
			for _, subfolder := range subfolders {
				subfolderIDs = append(subfolderIDs, subfolder.ID)
			}

			// Batch delete all categories for subfolders (1 query instead of N)
			if err := queries.BulkRemoveFolderCategories(ctx, subfolderIDs); err != nil {
				h.logger.Error("failed to bulk remove subfolder categories", zap.Error(err))
				return
			}

			// Prepare bulk insert data
			var folderIDs []pgtype.UUID
			var catIDs []pgtype.UUID
			for _, subfolderID := range subfolderIDs {
				for _, catUUID := range categoryUUIDs {
					folderIDs = append(folderIDs, subfolderID)
					catIDs = append(catIDs, catUUID)
				}
			}

			// Batch insert all categories (1 query instead of N×M)
			if len(folderIDs) > 0 {
				if err := queries.BulkAddFolderCategories(ctx, db.BulkAddFolderCategoriesParams{
					FolderIds:   folderIDs,
					CategoryIds: catIDs,
				}); err != nil {
					h.logger.Error("failed to bulk add subfolder categories", zap.Error(err))
					return
				}
			}

			h.logger.Info("bulk updated subfolder categories",
				zap.Int("subfolders", len(subfolders)),
				zap.Int("categories", len(categoryUUIDs)))

			// REMOVED: Recursive call - causes timeout on large folder structures
			// h.propagateFolderCategories(ctx, queries, subfolder.ID, categoryUUIDs, req)
		}
	}
}
