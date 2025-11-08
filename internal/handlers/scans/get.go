package scans

import (
	"net/http"

	"stl-manager/internal/db"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type ScanResponse struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	Found     int    `json:"found"`
	Processed int    `json:"processed"`
	Progress  int    `json:"progress"`
	Error     string `json:"error,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (h *Handler) GetScan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	scanID := chi.URLParam(r, "id")
	if scanID == "" {
		h.RespondError(w, http.StatusBadRequest, "scan_id is required")
		return
	}

	// Parse UUID
	uid, err := uuid.Parse(scanID)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, "invalid scan_id format")
		return
	}

	queries := db.New(h.pool)

	// Get scan from database
	scan, err := queries.GetScan(ctx, pgtype.UUID{Bytes: uid, Valid: true})
	if err != nil {
		h.logger.Error("failed to get scan", zap.String("scan_id", scanID), zap.Error(err))
		h.RespondError(w, http.StatusNotFound, "scan not found")
		return
	}

	// Build response
	response := ScanResponse{
		ID:        scanID,
		Status:    scan.Status,
		Found:     int(scan.Found.Int32),
		Processed: int(scan.Processed.Int32),
		Progress:  int(scan.Progress.Int32),
		CreatedAt: scan.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: scan.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}
	if scan.Error.Valid {
		response.Error = scan.Error.String
	}

	h.RespondJSON(w, http.StatusOK, response)
}
