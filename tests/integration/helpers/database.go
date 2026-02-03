package helpers

import (
	"context"
	"os"
	"testing"
	"time"

	"stl-manager/internal/database"
	"stl-manager/internal/db/sqlite"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var (
	// TestDB is the shared database connection for all tests
	TestDB database.Database
	// TestLogger is the shared logger for all tests
	TestLogger *zap.Logger
)

// SetupTestDatabase initializes the database connection for tests
// Call this from TestMain in your test packages
func SetupTestDatabase() error {
	// Setup logger
	var err error
	TestLogger, err = zap.NewDevelopment()
	if err != nil {
		return err
	}

	// Use in-memory SQLite database for tests
	dbPath := os.Getenv("TEST_DB_PATH")
	if dbPath == "" {
		dbPath = ":memory:"
	}

	TestDB, err = sqlite.Open(dbPath)
	if err != nil {
		TestLogger.Fatal("failed to open test database", zap.Error(err))
		return err
	}

	return nil
}

// CleanupTestDatabase closes the database connection
func CleanupTestDatabase() {
	if TestDB != nil {
		TestDB.Close()
	}
}

// Category Helpers

// CreateTestCategory creates a test category with a unique name
func CreateTestCategory(t *testing.T, name string) *database.Category {
	ctx := context.Background()

	// Add UUID to ensure uniqueness
	uniqueName := name + "-" + uuid.New().String()[:8]
	id := uuid.New().String()

	category, err := TestDB.CreateCategory(ctx, id, uniqueName)
	require.NoError(t, err, "Failed to create test category")

	return category
}

// DeleteTestCategory hard deletes a test category (cleanup)
func DeleteTestCategory(t *testing.T, id string) {
	ctx := context.Background()

	err := TestDB.SoftDeleteCategory(ctx, id)
	if err != nil {
		t.Logf("Warning: failed to delete test category: %v", err)
	}
}

// SoftDeleteTestCategory soft deletes a test category
func SoftDeleteTestCategory(t *testing.T, id string) {
	ctx := context.Background()

	err := TestDB.SoftDeleteCategory(ctx, id)
	require.NoError(t, err, "Failed to soft delete test category")
}

// RestoreTestCategory restores a soft deleted category
func RestoreTestCategory(t *testing.T, id string) {
	ctx := context.Background()

	err := TestDB.RestoreCategory(ctx, id)
	require.NoError(t, err, "Failed to restore test category")
}

// GetTestCategory retrieves a category by ID
func GetTestCategory(t *testing.T, id string) *database.Category {
	ctx := context.Background()

	category, err := TestDB.GetCategory(ctx, id)
	require.NoError(t, err, "Failed to get test category")

	return category
}

// File Helpers

// CreateTestFile creates a test file
func CreateTestFile(t *testing.T, fileName, fileType string, folderID *string) *database.File {
	ctx := context.Background()

	// Generate unique path
	id := uuid.New().String()
	uniquePath := "/test/path/" + fileName + "-" + uuid.New().String()[:8] + "." + fileType

	file, err := TestDB.UpsertFile(ctx, id, uniquePath, fileName+"."+fileType, fileType, 1024, time.Now(), nil, folderID)
	require.NoError(t, err, "Failed to create test file")

	return file
}

// GetTestFile retrieves a file by ID
func GetTestFile(t *testing.T, id string) *database.File {
	ctx := context.Background()

	file, err := TestDB.GetFile(ctx, id)
	require.NoError(t, err, "Failed to get test file")

	return file
}

// DeleteTestFile is a no-op for in-memory database
func DeleteTestFile(t *testing.T, id string) {
	// In-memory database cleans up automatically
}

// Folder Helpers

// CreateTestFolder creates a test folder
func CreateTestFolder(t *testing.T, name string) *database.Folder {
	ctx := context.Background()

	// Generate unique path
	id := uuid.New().String()
	uniquePath := "/test/folders/" + name + "-" + uuid.New().String()[:8]

	folder, err := TestDB.CreateFolder(ctx, id, name, uniquePath)
	require.NoError(t, err, "Failed to create test folder")

	return folder
}

// GetTestFolder retrieves a folder by ID
func GetTestFolder(t *testing.T, id string) *database.Folder {
	ctx := context.Background()

	folder, err := TestDB.GetFolder(ctx, id)
	require.NoError(t, err, "Failed to get test folder")

	return folder
}

// DeleteTestFolder is a no-op for in-memory database
func DeleteTestFolder(t *testing.T, id string) {
	// In-memory database cleans up automatically
}

// Scan Helpers

// CreateTestScan creates a test scan
func CreateTestScan(t *testing.T, status string) *database.Scan {
	ctx := context.Background()

	id := uuid.New().String()
	scan, err := TestDB.CreateScan(ctx, id, status, 0, 0, 0)
	require.NoError(t, err, "Failed to create test scan")

	return scan
}

// DeleteTestScan is a no-op for now (SQLite doesn't have DeleteScan in the interface)
func DeleteTestScan(t *testing.T, id string) {
	// SQLite implementation doesn't expose DeleteScan through the interface
	// Tests using in-memory database don't need explicit cleanup
}

// CreateTestScanPath creates a test scan path
func CreateTestScanPath(t *testing.T, scanID, rootPath string) *database.ScanPath {
	ctx := context.Background()

	id := uuid.New().String()
	scanPath, err := TestDB.CreateScanPath(ctx, id, scanID, rootPath)
	require.NoError(t, err, "Failed to create test scan path")

	return scanPath
}
