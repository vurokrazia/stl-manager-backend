package scans

import (
	"net/http"
	"strconv"

	"stl-manager/internal/db"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (h *Handler) ListScans(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := db.New(h.pool)

	// Parse pagination
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

	// Get scans from database
	scans, err := queries.ListScans(ctx, db.ListScansParams{
		Limit:  int32(pageSize),
		Offset: int32(offset),
	})
	if err != nil {
		h.logger.Error("failed to list scans", zap.Error(err))
		h.RespondError(w, http.StatusInternalServerError, "failed to list scans")
		return
	}

	// Get total count (FIX: was using len(scans) before)
	total, err := queries.CountScans(ctx)
	if err != nil {
		h.logger.Error("failed to count scans", zap.Error(err))
		total = 0
	}

	// Convert to response format
	items := make([]ScanResponse, len(scans))
	for i, scan := range scans {
		scanUUID := uuid.UUID(scan.ID.Bytes)
		items[i] = ScanResponse{
			ID:        scanUUID.String(),
			Status:    scan.Status,
			Found:     int(scan.Found.Int32),
			Processed: int(scan.Processed.Int32),
			Progress:  int(scan.Progress.Int32),
			CreatedAt: scan.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: scan.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		}
		if scan.Error.Valid {
			items[i].Error = scan.Error.String
		}
	}

	h.RespondJSON(w, http.StatusOK, map[string]any{
		"items":       items,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": int((total + int64(pageSize) - 1) / int64(pageSize)),
	})
}
