package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"stl-manager/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	files, err := filepath.Glob("migrations/*.sql")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read migrations: %v\n", err)
		os.Exit(1)
	}

	sort.Strings(files)

	for _, file := range files {
		fmt.Printf("Running: %s\n", filepath.Base(file))
		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read %s: %v\n", file, err)
			continue
		}

		_, err = pool.Exec(ctx, string(content))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to execute %s: %v\n", file, err)
			continue
		}

		fmt.Printf("✓ %s\n", filepath.Base(file))
	}

	fmt.Println("\nMigrations completed!")
}
