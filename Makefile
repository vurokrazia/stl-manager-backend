.PHONY: help dev build run migrate-up migrate-down sqlc test clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

dev: ## Run API server in development mode
	go run cmd/api/main.go

build: ## Build the API binary
	go build -o bin/stl-manager-api.exe cmd/api/main.go

run: build ## Build and run the API
	./bin/stl-manager-api.exe

migrate-up: ## Run database migrations (manual for now)
	@echo "Run migrations manually in Supabase SQL editor"
	@echo "File: internal/db/migrations/001_init.sql"

sqlc: ## Generate sqlc code
	sqlc generate

test: ## Run tests
	go test -v ./...

clean: ## Clean build artifacts
	rm -rf bin/
	go clean

deps: ## Install dependencies
	go mod download
	go mod tidy

fmt: ## Format code
	go fmt ./...

lint: ## Run linter
	golangci-lint run
