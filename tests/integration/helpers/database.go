package helpers

import (
	"context"
	"os"
	"testing"

	"stl-manager/internal/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var (
	// TestPool is the shared database connection pool for all tests
	TestPool *pgxpool.Pool
	// TestLogger is the shared logger for all tests
	TestLogger *zap.Logger
)

// SetupTestDatabase initializes the database connection for tests
// Call this from TestMain in your test packages
func SetupTestDatabase() error {
	// Load .env for test database connection
	_ = godotenv.Load("../../../.env")

	// Setup logger
	var err error
	TestLogger, err = zap.NewDevelopment()
	if err != nil {
		return err
	}

	// Connect to database
	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		TestLogger.Fatal("DATABASE_URL not set")
	}

	TestPool, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		TestLogger.Fatal("failed to connect to database", zap.Error(err))
	}

	return nil
}

// CleanupTestDatabase closes the database connection
func CleanupTestDatabase() {
	if TestPool != nil {
		TestPool.Close()
	}
}

// Category Helpers

// CreateTestCategory creates a test category with a unique name
func CreateTestCategory(t *testing.T, name string) *db.Category {
	ctx := context.Background()
	queries := db.New(TestPool)

	// Add UUID to ensure uniqueness
	uniqueName := name + "-" + uuid.New().String()[:8]

	category, err := queries.CreateCategory(ctx, uniqueName)
	require.NoError(t, err, "Failed to create test category")

	return &category
}

// DeleteTestCategory hard deletes a test category (cleanup)
func DeleteTestCategory(t *testing.T, id pgtype.UUID) {
	ctx := context.Background()
	queries := db.New(TestPool)

	err := queries.DeleteCategory(ctx, id)
	if err != nil {
		t.Logf("Warning: failed to delete test category: %v", err)
	}
}

// SoftDeleteTestCategory soft deletes a test category
func SoftDeleteTestCategory(t *testing.T, id pgtype.UUID) {
	ctx := context.Background()
	queries := db.New(TestPool)

	err := queries.SoftDeleteCategory(ctx, id)
	require.NoError(t, err, "Failed to soft delete test category")
}

// RestoreTestCategory restores a soft deleted category
func RestoreTestCategory(t *testing.T, id pgtype.UUID) {
	ctx := context.Background()
	queries := db.New(TestPool)

	err := queries.RestoreCategory(ctx, id)
	require.NoError(t, err, "Failed to restore test category")
}

// GetTestCategory retrieves a category by ID
func GetTestCategory(t *testing.T, id pgtype.UUID) *db.Category {
	ctx := context.Background()
	queries := db.New(TestPool)

	category, err := queries.GetCategory(ctx, id)
	require.NoError(t, err, "Failed to get test category")

	return &category
}

// File Helpers

// CreateTestFile creates a test file
func CreateTestFile(t *testing.T, fileName, fileType string, folderID pgtype.UUID) *db.File {
	ctx := context.Background()
	queries := db.New(TestPool)

	// Generate unique path
	uniquePath := "/test/path/" + fileName + "-" + uuid.New().String()[:8] + "." + fileType

	// Create file
	file, err := queries.CreateFile(ctx, db.CreateFileParams{
		Path:     uniquePath,
		FileName: fileName + "." + fileType,
		Type:     fileType,
		Size:     1024,
	})
	require.NoError(t, err, "Failed to create test file")

	// Update folder_id if provided
	if folderID.Valid {
		err = queries.UpdateFileFolderID(ctx, db.UpdateFileFolderIDParams{
			ID:       file.ID,
			FolderID: folderID,
		})
		require.NoError(t, err, "Failed to set file folder")

		// Refresh file to get updated folder_id
		file, err = queries.GetFile(ctx, file.ID)
		require.NoError(t, err, "Failed to get updated file")
	}

	return &file
}

// DeleteTestFile hard deletes a test file (cleanup)
func DeleteTestFile(t *testing.T, id pgtype.UUID) {
	ctx := context.Background()
	queries := db.New(TestPool)

	err := queries.DeleteFile(ctx, id)
	if err != nil {
		t.Logf("Warning: failed to delete test file: %v", err)
	}
}

// GetTestFile retrieves a file by ID
func GetTestFile(t *testing.T, id pgtype.UUID) *db.File {
	ctx := context.Background()
	queries := db.New(TestPool)

	file, err := queries.GetFile(ctx, id)
	require.NoError(t, err, "Failed to get test file")

	return &file
}

// Folder Helpers

// CreateTestFolder creates a test folder
func CreateTestFolder(t *testing.T, name string) *db.Folder {
	ctx := context.Background()
	queries := db.New(TestPool)

	// Generate unique path
	uniquePath := "/test/folders/" + name + "-" + uuid.New().String()[:8]

	folder, err := queries.CreateFolder(ctx, db.CreateFolderParams{
		Name: name,
		Path: uniquePath,
	})
	require.NoError(t, err, "Failed to create test folder")

	return &folder
}

// DeleteTestFolder hard deletes a test folder (cleanup)
func DeleteTestFolder(t *testing.T, id pgtype.UUID) {
	ctx := context.Background()
	queries := db.New(TestPool)

	err := queries.DeleteFolder(ctx, id)
	if err != nil {
		t.Logf("Warning: failed to delete test folder: %v", err)
	}
}

// GetTestFolder retrieves a folder by ID
func GetTestFolder(t *testing.T, id pgtype.UUID) *db.Folder {
	ctx := context.Background()
	queries := db.New(TestPool)

	folder, err := queries.GetFolder(ctx, id)
	require.NoError(t, err, "Failed to get test folder")

	return &folder
}

// Scan Helpers

// CreateTestScan creates a test scan
func CreateTestScan(t *testing.T, status string) *db.Scan {
	ctx := context.Background()
	queries := db.New(TestPool)

	scan, err := queries.CreateScan(ctx, db.CreateScanParams{
		Status:    status,
		Found:     pgtype.Int4{Int32: 0, Valid: true},
		Processed: pgtype.Int4{Int32: 0, Valid: true},
		Progress:  pgtype.Int4{Int32: 0, Valid: true},
	})
	require.NoError(t, err, "Failed to create test scan")

	return &scan
}

// DeleteTestScan hard deletes a test scan (cleanup)
func DeleteTestScan(t *testing.T, id pgtype.UUID) {
	ctx := context.Background()
	queries := db.New(TestPool)

	err := queries.DeleteScan(ctx, id)
	if err != nil {
		t.Logf("Warning: failed to delete test scan: %v", err)
	}
}
