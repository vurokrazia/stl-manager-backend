package main

import (
	"context"
	"fmt"
	"log"

	"stl-manager/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	fmt.Println("Applying migrations...")

	// Migration 1: Create folders table
	fmt.Println("\n[1/3] Creating folders table...")
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS folders (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT NOT NULL,
			path TEXT NOT NULL UNIQUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_folders_path ON folders(path);
		CREATE INDEX IF NOT EXISTS idx_folders_name ON folders(name);
	`)
	if err != nil {
		log.Fatalf("failed to create folders table: %v", err)
	}
	fmt.Println("✓ Folders table created")

	// Migration 2: Create folders_categories join table
	fmt.Println("\n[2/3] Creating folders_categories join table...")
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS folders_categories (
			folder_id UUID NOT NULL REFERENCES folders(id) ON DELETE CASCADE,
			category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
			PRIMARY KEY (folder_id, category_id)
		);

		CREATE INDEX IF NOT EXISTS idx_folders_categories_folder ON folders_categories(folder_id);
		CREATE INDEX IF NOT EXISTS idx_folders_categories_category ON folders_categories(category_id);
	`)
	if err != nil {
		log.Fatalf("failed to create folders_categories table: %v", err)
	}
	fmt.Println("✓ Folders_categories table created")

	// Migration 3: Add folder_id to files table
	fmt.Println("\n[3/3] Adding folder_id column to files table...")
	_, err = pool.Exec(ctx, `
		ALTER TABLE files ADD COLUMN IF NOT EXISTS folder_id UUID REFERENCES folders(id) ON DELETE SET NULL;

		CREATE INDEX IF NOT EXISTS idx_files_folder ON files(folder_id);
	`)
	if err != nil {
		log.Fatalf("failed to add folder_id to files: %v", err)
	}
	fmt.Println("✓ Folder_id column added to files")

	fmt.Println("\n✅ All migrations applied successfully!")
}
