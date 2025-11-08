package files

import (
	"encoding/json"
	"net/http"

	"stl-manager/internal/db"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

func (h *Handler) GetFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fileID := chi.URLParam(r, "id")
	if fileID == "" {
		h.RespondError(w, http.StatusBadRequest, "file_id is required")
		return
	}

	uid, err := uuid.Parse(fileID)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "invalid file_id format")
		return
	}

	queries := db.New(h.pool)
	file, err := queries.GetFile(ctx, pgtype.UUID{Bytes: uid, Valid: true})
	if err != nil {
		h.logger.Error("failed to get file", zap.String("file_id", fileID), zap.Error(err))
		h.RespondError(w, http.StatusNotFound, "file not found")
		return
	}

	categories, err := queries.GetFileCategories(ctx, file.ID)
	if err != nil {
		h.logger.Warn("failed to get file categories", zap.Error(err))
		categories = []db.Category{}
	}

	type FileWithCategories struct {
		db.File
		Categories []db.Category `json:"categories"`
	}

	h.RespondJSON(w, http.StatusOK, FileWithCategories{
		File:       file,
		Categories: categories,
	})
}

func (h *Handler) ReclassifyFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fileID := chi.URLParam(r, "id")
	if fileID == "" {
		h.RespondError(w, http.StatusBadRequest, "file_id is required")
		return
	}

	uid, err := uuid.Parse(fileID)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "invalid file_id format")
		return
	}

	queries := db.New(h.pool)
	file, err := queries.GetFile(ctx, pgtype.UUID{Bytes: uid, Valid: true})
	if err != nil {
		h.logger.Error("failed to get file", zap.String("file_id", fileID), zap.Error(err))
		h.RespondError(w, http.StatusNotFound, "file not found")
		return
	}

	allCategories, err := queries.ListCategories(ctx)
	if err != nil {
		h.logger.Error("failed to list categories", zap.Error(err))
		h.RespondError(w, http.StatusInternalServerError, "failed to fetch categories")
		return
	}

	if !h.classifier.IsEnabled() {
		h.RespondError(w, http.StatusServiceUnavailable, "OpenAI classification is not enabled")
		return
	}

	categoryNames := make([]string, len(allCategories))
	for i, cat := range allCategories {
		categoryNames[i] = cat.Name
	}

	classifiedCategories, err := h.classifier.Classify(ctx, file.FileName, categoryNames)
	if err != nil {
		h.logger.Error("classification failed", zap.String("file_id", fileID), zap.Error(err))
		h.RespondError(w, http.StatusInternalServerError, "classification failed")
		return
	}

	if len(classifiedCategories) == 0 {
		classifiedCategories = []string{"uncategorized"}
	}

	err = queries.RemoveAllFileCategories(ctx, file.ID)
	if err != nil {
		h.logger.Error("failed to remove existing categories", zap.String("file_id", fileID), zap.Error(err))
	}

	for _, catName := range classifiedCategories {
		category, err := queries.GetCategoryByName(ctx, catName)
		if err != nil {
			h.logger.Warn("category not found, skipping", zap.String("category", catName))
			continue
		}

		err = queries.AddFileCategory(ctx, db.AddFileCategoryParams{
			FileID:     file.ID,
			CategoryID: category.ID,
		})
		if err != nil {
			h.logger.Error("failed to add category",
				zap.String("file_id", fileID),
				zap.String("category", catName),
				zap.Error(err))
		}
	}

	h.logger.Info("file reclassified",
		zap.String("file_id", fileID),
		zap.Strings("categories", classifiedCategories))

	h.RespondJSON(w, http.StatusOK, map[string]any{
		"file_id":    fileID,
		"categories": classifiedCategories,
	})
}

type UpdateCategoriesRequest struct {
	CategoryIDs []string `json:"category_ids"`
}

func (h *Handler) UpdateFileCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fileID := chi.URLParam(r, "id")
	if fileID == "" {
		h.RespondError(w, http.StatusBadRequest, "file_id is required")
		return
	}

	uid, err := uuid.Parse(fileID)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "invalid file_id format")
		return
	}

	queries := db.New(h.pool)
	file, err := queries.GetFile(ctx, pgtype.UUID{Bytes: uid, Valid: true})
	if err != nil {
		h.logger.Error("failed to get file", zap.String("file_id", fileID), zap.Error(err))
		h.RespondError(w, http.StatusNotFound, "file not found")
		return
	}

	var req UpdateCategoriesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = queries.RemoveAllFileCategories(ctx, file.ID)
	if err != nil {
		h.logger.Error("failed to remove existing categories", zap.String("file_id", fileID), zap.Error(err))
		h.RespondError(w, http.StatusInternalServerError, "failed to update categories")
		return
	}

	for _, catIDStr := range req.CategoryIDs {
		catUID, err := uuid.Parse(catIDStr)
		if err != nil {
			h.logger.Warn("invalid category ID, skipping", zap.String("category_id", catIDStr))
			continue
		}

		err = queries.AddFileCategory(ctx, db.AddFileCategoryParams{
			FileID:     file.ID,
			CategoryID: pgtype.UUID{Bytes: catUID, Valid: true},
		})
		if err != nil {
			h.logger.Error("failed to add category",
				zap.String("file_id", fileID),
				zap.String("category_id", catIDStr),
				zap.Error(err))
		}
	}

	categories, err := queries.GetFileCategories(ctx, file.ID)
	if err != nil {
		h.logger.Error("failed to get updated categories", zap.Error(err))
		categories = []db.Category{}
	}

	h.logger.Info("file categories updated",
		zap.String("file_id", fileID),
		zap.Int("category_count", len(categories)))

	h.RespondJSON(w, http.StatusOK, map[string]any{
		"file_id":    fileID,
		"categories": categories,
	})
}
