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
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer pool.Close()

	// Count root files
	var rootCount int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM files WHERE folder_id IS NULL").Scan(&rootCount)
	if err != nil {
		log.Fatal("Failed to count root files:", err)
	}

	// Count files with folder
	var folderCount int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM files WHERE folder_id IS NOT NULL").Scan(&folderCount)
	if err != nil {
		log.Fatal("Failed to count folder files:", err)
	}

	// Count folders
	var foldersCount int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM folders").Scan(&foldersCount)
	if err != nil {
		log.Fatal("Failed to count folders:", err)
	}

	fmt.Printf("Root-level files (folder_id IS NULL): %d\n", rootCount)
	fmt.Printf("Files inside folders (folder_id IS NOT NULL): %d\n", folderCount)
	fmt.Printf("Total folders: %d\n", foldersCount)
	fmt.Printf("Total files: %d\n", rootCount+folderCount)
}
