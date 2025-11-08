package handlers

import (
	"encoding/json"
	"net/http"

	"stl-manager/internal/ai"
	"stl-manager/internal/config"
	"stl-manager/internal/scanner"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Handler holds shared dependencies for all handlers
type Handler struct {
	pool       *pgxpool.Pool
	classifier ai.Classifier
	scanner    *scanner.Scanner
	config     *config.Config
	logger     *zap.Logger
}

// New creates a new Handler instance
func New(pool *pgxpool.Pool, classifier ai.Classifier, scanner *scanner.Scanner, cfg *config.Config, logger *zap.Logger) *Handler {
	return &Handler{
		pool:       pool,
		classifier: classifier,
		scanner:    scanner,
		config:     cfg,
		logger:     logger,
	}
}

// Response helpers

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}
