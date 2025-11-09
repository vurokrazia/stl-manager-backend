.PHONY: help dev build run migrate-up migrate-down sqlc test test-integration test-coverage test-ci clean fmt lint ci-setup

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

dev: ## Run API server in development mode
	go run cmd/api/main.go

build: ## Build the API binary
	go build -o bin/stl-manager-api cmd/api/main.go

run: build ## Build and run the API
	./bin/stl-manager-api

migrate-up: ## Run database migrations
	@echo "Running migrations..."
	@for migration in migrations/*.sql; do \
		echo "Applying $$migration..."; \
		psql $(DATABASE_URL) -f "$$migration"; \
	done

migrate-down: ## Rollback database migrations (not implemented)
	@echo "Migration rollback not implemented yet"

sqlc: ## Generate sqlc code
	sqlc generate

# Test commands
test: ## Run all tests
	go test -v ./...

test-integration: ## Run integration tests only
	go test -v ./tests/integration/...

test-unit: ## Run unit tests only (exclude integration tests)
	go test -v $$(go list ./... | grep -v /tests/)

test-coverage: ## Run tests with coverage report
	go test -v ./tests/integration/... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-ci: ## Run tests for CI (with race detection)
	go test -v ./tests/integration/... -race -coverprofile=coverage.txt -covermode=atomic

test-watch: ## Run tests in watch mode (requires entr)
	ls **/*.go | entr -c go test ./tests/integration/... -v

# Quality checks
clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html coverage.txt
	go clean

deps: ## Install dependencies
	go mod download
	go mod tidy

fmt: ## Format code
	go fmt ./...
	goimports -w .

fmt-check: ## Check code formatting
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "The following files are not formatted:"; \
		gofmt -l .; \
		echo "Please run: make fmt"; \
		exit 1; \
	fi

lint: ## Run linter
	golangci-lint run --timeout=5m

lint-fix: ## Run linter and fix issues
	golangci-lint run --fix --timeout=5m

# CI/CD helpers
ci-setup: deps ## Setup for CI environment
	@echo "Setting up CI environment..."
	go mod verify

ci-test: fmt-check lint test-ci ## Run all CI checks

# Development helpers
watch: ## Watch and rebuild on changes (requires entr)
	ls **/*.go | entr -r go run cmd/api/main.go

install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install golang.org/x/tools/cmd/goimports@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin

docker-build: ## Build Docker image
	docker build -t stl-manager-api .

docker-run: ## Run Docker container
	docker run -p 8080:8080 --env-file .env stl-manager-api
