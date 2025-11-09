package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Get database URL
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://stl_user:stl_password@localhost:5432/stl_manager?sslmode=disable"
	}

	// Connect to database
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	// Run migration
	migration := `
ALTER TABLE folders
ADD COLUMN parent_folder_id UUID REFERENCES folders(id) ON DELETE CASCADE;

CREATE INDEX idx_folders_parent_folder_id ON folders(parent_folder_id);
`

	fmt.Println("Running migration 007: Add parent_folder_id to folders table...")

	_, err = pool.Exec(ctx, migration)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("Migration 007 completed successfully!")
}
