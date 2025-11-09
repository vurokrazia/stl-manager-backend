package scans

import (
	"context"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"stl-manager/internal/db"
	"stl-manager/internal/scanner"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

// runScan executes the scan process
func (h *Handler) runScan(ctx context.Context, scanID uuid.UUID) {
	h.logger.Info("running scan", zap.String("scan_id", scanID.String()))
	queries := db.New(h.pool)
	scanUUID := pgtype.UUID{Bytes: scanID, Valid: true}

	// Update scan status to running
	updateScanStatus := func(status string, found, processed, progress int, errorMsg string) {
		_, err := queries.UpdateScan(ctx, db.UpdateScanParams{
			ID:        scanUUID,
			Status:    status,
			Found:     pgtype.Int4{Int32: int32(found), Valid: true},
			Processed: pgtype.Int4{Int32: int32(processed), Valid: true},
			Progress:  pgtype.Int4{Int32: int32(progress), Valid: true},
			Error:     pgtype.Text{String: errorMsg, Valid: errorMsg != ""},
		})
		if err != nil {
			h.logger.Error("failed to update scan status", zap.Error(err))
		}
	}

	// Scan files
	files, err := h.scanner.Scan(ctx)
	if err != nil {
		h.logger.Error("scan failed", zap.Error(err))
		updateScanStatus("failed", 0, 0, 0, err.Error())
		return
	}

	h.logger.Info("scan completed",
		zap.String("scan_id", scanID.String()),
		zap.Int("files_found", len(files)),
	)

	// Update scan with found count
	updateScanStatus("running", len(files), 0, 5, "")

	// PHASE 1: Discover and create complete folder hierarchy (only folders with files)
	h.logger.Info("discovering folder hierarchy")
	folderCache, err := h.discoverAndCreateFolderHierarchy(ctx, queries, files)
	if err != nil {
		h.logger.Error("failed to create folder hierarchy", zap.Error(err))
		updateScanStatus("failed", len(files), 0, 0, err.Error())
		return
	}
	h.logger.Info("folder hierarchy created", zap.Int("total_folders", len(folderCache)))
	updateScanStatus("running", len(files), 0, 10, "")

	// Get all categories for classification
	allCategories, err := queries.ListCategories(ctx)
	if err != nil {
		h.logger.Error("failed to list categories for classification", zap.Error(err))
		allCategories = []db.Category{}
	}

	categoryNames := make([]string, len(allCategories))
	categoryMap := make(map[string]pgtype.UUID)
	for i, cat := range allCategories {
		categoryNames[i] = cat.Name
		categoryMap[cat.Name] = cat.ID
	}

	// PHASE 2: Process files in parallel with controlled concurrency
	var (
		processed    = 0
		mu           sync.Mutex
		wg           sync.WaitGroup
		sem          = make(chan struct{}, 20) // Max 20 concurrent workers
		progressChan = make(chan int, len(files))
	)

	// Progress updater goroutine
	go func() {
		lastUpdate := 0
		for range progressChan {
			mu.Lock()
			processed++
			current := processed
			mu.Unlock()

			// Update progress every 50 files or at completion
			if current%50 == 0 || current == len(files) {
				progress := 10 + int(float64(current)/float64(len(files))*80)
				updateScanStatus("running", len(files), current, progress, "")
				h.logger.Info("scan progress",
					zap.String("scan_id", scanID.String()),
					zap.Int("processed", current),
					zap.Int("total", len(files)),
					zap.Int("progress", progress))
				lastUpdate = current
			}
		}
		// Final update if not aligned with batch
		if lastUpdate != len(files) {
			mu.Lock()
			current := processed
			mu.Unlock()
			progress := 10 + int(float64(current)/float64(len(files))*80)
			updateScanStatus("running", len(files), current, progress, "")
		}
	}()

	// Process files in parallel
	for _, file := range files {
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore

		go func(f scanner.FileInfo) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			// Get folder ID from cache
			var folderID pgtype.UUID
			if f.FolderPath != "" {
				if cachedID, ok := folderCache[f.FolderPath]; ok {
					folderID = cachedID
				} else {
					h.logger.Warn("folder not found in cache",
						zap.String("path", f.FolderPath))
				}
			}

			// Upsert file
			savedFile, err := queries.UpsertFile(ctx, db.UpsertFileParams{
				Path:       f.Path,
				FileName:   f.FileName,
				Type:       f.Type,
				Size:       f.Size,
				ModifiedAt: pgtype.Timestamptz{Time: f.ModifiedAt, Valid: true},
				Sha256:     pgtype.Text{String: f.SHA256, Valid: f.SHA256 != ""},
				FolderID:   folderID,
			})

			if err != nil {
				h.logger.Error("failed to save file",
					zap.String("path", f.Path),
					zap.Error(err))
				progressChan <- 1
				return
			}

			// Classify file with OpenAI
			var classifiedCategories []string
			if h.classifier.IsEnabled() {
				classifiedCategories, err = h.classifier.Classify(ctx, f.FileName, categoryNames)
				if err != nil {
					h.logger.Warn("classification failed",
						zap.String("file", f.FileName),
						zap.Error(err))
					classifiedCategories = []string{"uncategorized"}
				}
				if len(classifiedCategories) == 0 {
					classifiedCategories = []string{"uncategorized"}
				}
			} else {
				classifiedCategories = []string{"uncategorized"}
			}

			// Remove existing categories and add new ones
			_ = queries.RemoveAllFileCategories(ctx, savedFile.ID)
			for _, catName := range classifiedCategories {
				if catID, ok := categoryMap[catName]; ok {
					err = queries.AddFileCategory(ctx, db.AddFileCategoryParams{
						FileID:     savedFile.ID,
						CategoryID: catID,
					})
					if err != nil {
						h.logger.Error("failed to add category",
							zap.String("file", f.FileName),
							zap.String("category", catName),
							zap.Error(err))
					}
				}
			}

			h.logger.Debug("saved and classified file",
				zap.String("path", f.Path),
				zap.Strings("categories", classifiedCategories))

			progressChan <- 1
		}(file)
	}

	// Wait for all workers to finish
	wg.Wait()
	close(progressChan)

	// Wait for progress updater to finish
	time.Sleep(100 * time.Millisecond)

	// Mark scan as completed
	updateScanStatus("completed", len(files), processed, 100, "")
	h.logger.Info("scan completed successfully",
		zap.String("scan_id", scanID.String()),
		zap.Int("files_processed", processed))
}

// discoverAndCreateFolderHierarchy discovers folders that contain files (or are ancestors of such folders)
// and creates them with proper parent_folder_id relationships. Empty folders are NOT registered.
func (h *Handler) discoverAndCreateFolderHierarchy(ctx context.Context, queries *db.Queries, files []scanner.FileInfo) (map[string]pgtype.UUID, error) {
	rootDir := h.config.ScanRootDir
	folderCache := make(map[string]pgtype.UUID)

	// Helper to get folder info from path
	getParentPath := func(fullPath string) (parentPath string, hasParent bool) {
		dir := filepath.Dir(fullPath)
		cleanRoot := filepath.Clean(rootDir)
		cleanDir := filepath.Clean(dir)

		if cleanDir == cleanRoot || cleanDir == "." {
			return "", false
		}
		return cleanDir, true
	}

	// Build a set of all folder paths that contain files (directly or indirectly)
	foldersToCreate := make(map[string]bool)

	for _, file := range files {
		if file.FolderPath == "" {
			continue // File at root level, no folder needed
		}

		// Add the immediate parent folder
		folderPath := filepath.Clean(file.FolderPath)
		foldersToCreate[folderPath] = true

		// Add all ancestor folders up to the root
		currentPath := folderPath
		for {
			parentPath, hasParent := getParentPath(currentPath)
			if !hasParent {
				break
			}
			foldersToCreate[parentPath] = true
			currentPath = parentPath
		}
	}

	// Convert map to slice
	var allFolders []string
	for folderPath := range foldersToCreate {
		allFolders = append(allFolders, folderPath)
	}

	h.logger.Info("discovered folders with files", zap.Int("count", len(allFolders)))

	// Sort folders by depth (shallowest first) to ensure parents are created before children
	sort.Slice(allFolders, func(i, j int) bool {
		depthI := strings.Count(allFolders[i], string(filepath.Separator))
		depthJ := strings.Count(allFolders[j], string(filepath.Separator))
		return depthI < depthJ
	})

	// Create folders in order (parents first, then children)
	for _, folderPath := range allFolders {
		// Determine parent_folder_id first (before checking if folder exists)
		var parentFolderID pgtype.UUID
		parentPath, hasParent := getParentPath(folderPath)
		if hasParent {
			if parentID, ok := folderCache[parentPath]; ok {
				parentFolderID = parentID
			} else {
				h.logger.Warn("parent folder not in cache (order issue?)",
					zap.String("folder", folderPath),
					zap.String("parent", parentPath))
			}
		}

		// Check if already exists in DB
		existing, err := queries.GetFolderByPath(ctx, folderPath)
		if err == nil {
			// Folder exists - update its parent_folder_id if needed
			if existing.ParentFolderID != parentFolderID {
				h.logger.Debug("updating folder parent",
					zap.String("folder", folderPath),
					zap.Bool("has_parent", hasParent))

				_, err := queries.UpdateFolderParent(ctx, db.UpdateFolderParentParams{
					ID:             existing.ID,
					ParentFolderID: parentFolderID,
				})
				if err != nil {
					h.logger.Error("failed to update folder parent",
						zap.String("path", folderPath),
						zap.Error(err))
				}
			}
			folderCache[folderPath] = existing.ID
			continue
		}

		// Create folder with parent_folder_id
		folderName := filepath.Base(folderPath)

		// Use CreateFolderWithParent if parent exists, otherwise CreateFolder
		var created db.Folder
		if hasParent && parentFolderID.Valid {
			created, err = queries.CreateFolderWithParent(ctx, db.CreateFolderWithParentParams{
				Name:           folderName,
				Path:           folderPath,
				ParentFolderID: parentFolderID,
			})
		} else {
			created, err = queries.CreateFolder(ctx, db.CreateFolderParams{
				Name: folderName,
				Path: folderPath,
			})
		}

		if err != nil {
			h.logger.Error("failed to create folder",
				zap.String("path", folderPath),
				zap.Error(err))
			continue
		}

		folderCache[folderPath] = created.ID
		h.logger.Debug("created folder",
			zap.String("name", folderName),
			zap.String("path", folderPath),
			zap.Bool("has_parent", hasParent))
	}

	return folderCache, nil
}
