package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"stl-manager/internal/ai"
	"stl-manager/internal/config"
	"stl-manager/internal/handlers"
	"stl-manager/internal/handlers/browse"
	"stl-manager/internal/handlers/categories"
	"stl-manager/internal/handlers/files"
	"stl-manager/internal/handlers/folders"
	"stl-manager/internal/handlers/scans"
	"stl-manager/internal/scanner"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = logger.Sync() }()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	logger.Info("starting stl-manager API",
		zap.String("port", cfg.Port),
		zap.String("scan_root", cfg.ScanRootDir),
	)

	// Initialize database connection
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		logger.Fatal("failed to ping database", zap.Error(err))
	}

	logger.Info("connected to database")

	// Initialize services
	classifier := ai.NewOpenAIClassifier(cfg.OpenAIAPIKey)
	fileScanner := scanner.New(cfg.ScanRootDir, cfg.SupportedExts, logger)

	// Initialize modular handlers
	baseHandler := handlers.New(pool, classifier, fileScanner, cfg, logger)
	scansHandler := scans.New(pool, classifier, fileScanner, cfg, logger)
	filesHandler := files.New(pool, classifier, cfg, logger)
	foldersHandler := folders.New(pool, logger)
	categoriesHandler := categories.New(pool, logger)
	browseHandler := browse.New(pool, logger)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// API Key middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" || apiKey != cfg.APIKey {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// Routes
	r.Route("/v1", func(r chi.Router) {
		// Health check
		r.Get("/health", baseHandler.Health)

		// Scans
		r.Post("/scan", scansHandler.CreateScan)
		r.Get("/scans/{id}", scansHandler.GetScan)
		r.Get("/scans", scansHandler.ListScans)

		// Files
		r.Get("/files", filesHandler.ListFiles)
		r.Get("/files/{id}", filesHandler.GetFile)
		r.Post("/files/{id}/reclassify", filesHandler.ReclassifyFile)
		r.Patch("/files/{id}/categories", filesHandler.UpdateFileCategories)

		// Categories
		r.Get("/categories", categoriesHandler.ListCategories)
		r.Post("/categories", categoriesHandler.CreateCategory)
		r.Get("/categories/{id}", categoriesHandler.GetCategory)
		r.Put("/categories/{id}", categoriesHandler.UpdateCategory)
		r.Delete("/categories/{id}", categoriesHandler.SoftDeleteCategory)
		r.Post("/categories/{id}/restore", categoriesHandler.RestoreCategory)

		// Browse - Mixed view of folders and root files
		r.Get("/browse", browseHandler.ListBrowse)

		// Mixed - ONLY folders and root-level files (dedicated endpoint)
		r.Get("/mixed", browseHandler.ListMixed)

		// Folders
		r.Get("/folders", foldersHandler.ListFolders)
		r.Get("/folders/{id}", foldersHandler.GetFolder)
		r.Patch("/folders/{id}/categories", foldersHandler.UpdateFolderCategories)

		// AI
		r.Get("/ai/status", baseHandler.GetAIStatus)
	})

	// Start server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		logger.Info("shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("server shutdown error", zap.Error(err))
		}
	}()

	logger.Info("server started", zap.String("addr", srv.Addr))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("server failed", zap.Error(err))
	}

	logger.Info("server stopped")
}
