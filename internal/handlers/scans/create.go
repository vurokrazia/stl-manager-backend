package scans

import (
	"context"
	"net/http"

	"stl-manager/internal/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type CreateScanResponse struct {
	ScanID string `json:"scan_id"`
}

func (h *Handler) CreateScan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := db.New(h.Pool())

	// Create scan record in database
	scan, err := queries.CreateScan(ctx, db.CreateScanParams{
		Status:    "running",
		Found:     pgtype.Int4{Int32: 0, Valid: true},
		Processed: pgtype.Int4{Int32: 0, Valid: true},
		Progress:  pgtype.Int4{Int32: 0, Valid: true},
	})
	if err != nil {
		h.Logger().Error("failed to create scan record", zap.Error(err))
		h.RespondError(w, http.StatusInternalServerError, "failed to create scan")
		return
	}

	scanUUID := uuid.UUID(scan.ID.Bytes)
	h.Logger().Info("scan started", zap.String("scan_id", scanUUID.String()))

	// Start scan in goroutine
	go h.runScan(context.Background(), scanUUID)

	h.RespondJSON(w, http.StatusAccepted, CreateScanResponse{
		ScanID: scanUUID.String(),
	})
}
