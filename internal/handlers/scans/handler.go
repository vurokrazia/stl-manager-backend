package scans

import (
	"encoding/json"
	"net/http"

	"stl-manager/internal/ai"
	"stl-manager/internal/config"
	"stl-manager/internal/scanner"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Handler wraps the base handler dependencies for scans
type Handler struct {
	pool       *pgxpool.Pool
	classifier ai.Classifier
	scanner    *scanner.Scanner
	config     *config.Config
	logger     *zap.Logger
}

// New creates a new scans Handler
func New(pool *pgxpool.Pool, classifier ai.Classifier, scanner *scanner.Scanner, cfg *config.Config, logger *zap.Logger) *Handler {
	return &Handler{
		pool:       pool,
		classifier: classifier,
		scanner:    scanner,
		config:     cfg,
		logger:     logger,
	}
}

// Getters for dependencies
func (h *Handler) Pool() *pgxpool.Pool {
	return h.pool
}

func (h *Handler) Classifier() ai.Classifier {
	return h.classifier
}

func (h *Handler) Scanner() *scanner.Scanner {
	return h.scanner
}

func (h *Handler) Config() *config.Config {
	return h.config
}

func (h *Handler) Logger() *zap.Logger {
	return h.logger
}

// Response helpers
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
