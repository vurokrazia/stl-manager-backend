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

// File Helpers (for future use)

// CreateTestFile creates a test file
// func CreateTestFile(t *testing.T, params db.CreateFileParams) *db.File {
// 	ctx := context.Background()
// 	queries := db.New(TestPool)
//
// 	file, err := queries.CreateFile(ctx, params)
// 	require.NoError(t, err, "Failed to create test file")
//
// 	return &file
// }

// Folder Helpers (for future use)

// CreateTestFolder creates a test folder
// func CreateTestFolder(t *testing.T, name string, path string) *db.Folder {
// 	ctx := context.Background()
// 	queries := db.New(TestPool)
//
// 	folder, err := queries.CreateFolder(ctx, db.CreateFolderParams{
// 		Name: name,
// 		Path: path,
// 	})
// 	require.NoError(t, err, "Failed to create test folder")
//
// 	return &folder
// }
