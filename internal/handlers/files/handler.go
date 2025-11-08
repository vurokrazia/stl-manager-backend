package files

import (
	"encoding/json"
	"net/http"

	"stl-manager/internal/ai"
	"stl-manager/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Handler struct {
	pool       *pgxpool.Pool
	classifier ai.Classifier
	config     *config.Config
	logger     *zap.Logger
}

func New(pool *pgxpool.Pool, classifier ai.Classifier, cfg *config.Config, logger *zap.Logger) *Handler {
	return &Handler{
		pool:       pool,
		classifier: classifier,
		config:     cfg,
		logger:     logger,
	}
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
